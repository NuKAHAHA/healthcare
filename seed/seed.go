// Seed script — DEVELOPMENT ONLY.
// Creates 25 users, 200 patients, 300 appointments, 250 treatments.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"healthcare-api/internal/auth"
	"healthcare-api/internal/config"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/repositories"
	"healthcare-api/internal/services"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// ── Environment guard ────────────────────────────────────
	if os.Getenv("APP_ENV") != "development" {
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════════════════╗")
		fmt.Println("║  SEED REFUSED — APP_ENV must be 'development'           ║")
		fmt.Println("║  Set APP_ENV=development before running this script.    ║")
		fmt.Println("╚══════════════════════════════════════════════════════════╝")
		fmt.Println()
		log.Fatal("Seed aborted: production guard triggered")
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  WARNING: DEVELOPMENT SEED                              ║")
	fmt.Println("║  This will INSERT test data into your database.         ║")
	fmt.Println("║  NEVER run this script in production!                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v\n", err)
	}

	appLogger := logger.New("info")
	appLogger.Info("Starting bulk seed")

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.GetDSN())
	if err != nil {
		log.Fatalf("db connect: %v\n", err)
	}
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	refreshRepo := repositories.NewRefreshTokenRepository(db)
	passwordMgr := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager(&cfg.JWT)
	authService := services.NewAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)

	// ── 1. Users (25 total) ──────────────────────────────────
	adminPwd := mustRandomHex(16) // printed once to console — never stored in code
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Generated admin password: %-32s║\n", adminPwd)
	fmt.Printf("║  Save it now — it will not be shown again.              ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n\n")

	users := buildUserList(adminPwd)
	createdUsers := make(map[string]int64) // role → first user of that role ID

	for _, u := range users {
		created, err := authService.CreateUser(ctx, u.email, u.password, u.firstName, u.lastName, u.role)
		if err != nil {
			appLogger.WarnWithContext(fmt.Sprintf("skip user %s: %v", u.email, err), "seed", 0, "")
			continue
		}
		appLogger.InfoWithContext(fmt.Sprintf("user %s (%s)", u.email, u.role), "seed", 0, "")
		if _, seen := createdUsers[u.role]; !seen {
			createdUsers[u.role] = created.ID
		}
	}

	// Collect doctor IDs for appointment assignment
	doctorIDs, err := getDoctorIDs(ctx, db)
	if err != nil || len(doctorIDs) == 0 {
		log.Fatalf("no doctors found after seeding users: %v", err)
	}
	registrarID := createdUsers["registrar"]
	if registrarID == 0 {
		log.Fatal("no registrar created")
	}

	// ── 2. Patients (200) ────────────────────────────────────
	patientIDs, err := insertPatients(ctx, db, registrarID, 200)
	if err != nil {
		log.Fatalf("insert patients: %v", err)
	}
	appLogger.InfoWithContext(fmt.Sprintf("inserted %d patients", len(patientIDs)), "seed", 0, "")

	// ── 3. Appointments (300) ────────────────────────────────
	// 250 completed + 30 pending + 20 cancelled
	apptIDs, err := insertAppointments(ctx, db, patientIDs, doctorIDs)
	if err != nil {
		log.Fatalf("insert appointments: %v", err)
	}
	appLogger.InfoWithContext(fmt.Sprintf("inserted %d appointments", len(apptIDs.completed)+len(apptIDs.pending)+len(apptIDs.cancelled)), "seed", 0, "")

	// ── 4. Treatments (250) for completed appointments ───────
	count, err := insertTreatments(ctx, db, apptIDs.completed)
	if err != nil {
		log.Fatalf("insert treatments: %v", err)
	}
	appLogger.InfoWithContext(fmt.Sprintf("inserted %d treatments", count), "seed", 0, "")

	fmt.Println()
	fmt.Println("Seed complete!")
	fmt.Printf("  Users:        %d\n", len(users))
	fmt.Printf("  Patients:     %d\n", len(patientIDs))
	fmt.Printf("  Appointments: %d\n", len(apptIDs.completed)+len(apptIDs.pending)+len(apptIDs.cancelled))
	fmt.Printf("  Treatments:   %d\n", count)
	fmt.Println()
	fmt.Printf("Admin:     admin@healthcare.local / %s\n", adminPwd)
	fmt.Printf("Registrar: registrar1@healthcare.local / Registrar@seed2024\n")
	fmt.Printf("Doctor:    doctor1@healthcare.local  / Doctor@seed2024\n")
}

// ── data builders ────────────────────────────────────────────────────────────

type seedUser struct{ email, password, firstName, lastName, role string }

func buildUserList(adminPwd string) []seedUser {
	list := []seedUser{
		{"admin@healthcare.local", adminPwd, "System", "Admin", "admin"},
	}
	for i := 1; i <= 5; i++ {
		list = append(list, seedUser{
			fmt.Sprintf("registrar%d@healthcare.local", i),
			"Registrar@seed2024",
			registrarFirstNames[i-1], "Registrar", "registrar",
		})
	}
	for i := 1; i <= 19; i++ {
		list = append(list, seedUser{
			fmt.Sprintf("doctor%d@healthcare.local", i),
			"Doctor@seed2024",
			doctorFirstNames[(i-1)%len(doctorFirstNames)],
			doctorLastNames[(i-1)%len(doctorLastNames)],
			"doctor",
		})
	}
	return list
}

func insertPatients(ctx context.Context, db *pgxpool.Pool, registrarID int64, n int) ([]int64, error) {
	genders := []string{"M", "F", "O"}
	ids := make([]int64, 0, n)

	for i := 0; i < n; i++ {
		fn := firstNames[i%len(firstNames)]
		ln := lastNames[(i*7)%len(lastNames)]
		email := fmt.Sprintf("%s.%s.%d@example.com", fn, ln, i)
		phone := fmt.Sprintf("+1202555%04d", i+1000)
		dob := time.Date(1950+randN(51), time.Month(1+randN(12)), 1+randN(28), 0, 0, 0, 0, time.UTC)
		gender := genders[i%3]

		var id int64
		err := db.QueryRow(ctx, `
			INSERT INTO patients
			  (first_name, last_name, email, phone, date_of_birth, gender, address, medical_info, registered_by, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())
			RETURNING id`,
			fn, ln, email, phone, dob, gender,
			fmt.Sprintf("%d Medical Street, Health City", 100+i),
			medicalInfos[i%len(medicalInfos)],
			registrarID,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("patient %d: %w", i, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

type apptSets struct{ completed, pending, cancelled []int64 }

func insertAppointments(ctx context.Context, db *pgxpool.Pool, patientIDs, doctorIDs []int64) (apptSets, error) {
	var sets apptSets
	now := time.Now()

	for i := 0; i < 300; i++ {
		patientID := patientIDs[i%len(patientIDs)]
		doctorID := doctorIDs[i%len(doctorIDs)]
		reason := appointmentReasons[i%len(appointmentReasons)]

		var status string
		var scheduledAt time.Time

		switch {
		case i < 250: // completed — past dates
			status = "completed"
			daysAgo := 1 + randN(365)
			scheduledAt = now.Add(-time.Duration(daysAgo) * 24 * time.Hour)
		case i < 280: // pending — future dates
			status = "pending"
			daysAhead := 1 + randN(90)
			scheduledAt = now.Add(time.Duration(daysAhead) * 24 * time.Hour)
		default: // cancelled — past dates
			status = "cancelled"
			daysAgo := 1 + randN(180)
			scheduledAt = now.Add(-time.Duration(daysAgo) * 24 * time.Hour)
		}

		var id int64
		err := db.QueryRow(ctx, `
			INSERT INTO appointments
			  (patient_id, doctor_id, scheduled_at, reason, status, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,NOW(),NOW())
			RETURNING id`,
			patientID, doctorID, scheduledAt, reason, status,
		).Scan(&id)
		if err != nil {
			return sets, fmt.Errorf("appointment %d: %w", i, err)
		}

		switch status {
		case "completed":
			sets.completed = append(sets.completed, id)
		case "pending":
			sets.pending = append(sets.pending, id)
		case "cancelled":
			sets.cancelled = append(sets.cancelled, id)
		}
	}
	return sets, nil
}

func insertTreatments(ctx context.Context, db *pgxpool.Pool, completedApptIDs []int64) (int, error) {
	count := 0
	// Insert treatments for at most 250 completed appointments
	limit := 250
	if len(completedApptIDs) < limit {
		limit = len(completedApptIDs)
	}

	for i := 0; i < limit; i++ {
		apptID := completedApptIDs[i]

		// Look up patient_id and doctor_id for this appointment
		var patientID, doctorID int64
		err := db.QueryRow(ctx,
			`SELECT patient_id, doctor_id FROM appointments WHERE id = $1`, apptID,
		).Scan(&patientID, &doctorID)
		if err != nil {
			return count, fmt.Errorf("lookup appt %d: %w", apptID, err)
		}

		diagnosis := diagnoses[i%len(diagnoses)]
		prescription := prescriptions[i%len(prescriptions)]

		_, err = db.Exec(ctx, `
			INSERT INTO treatments
			  (appointment_id, patient_id, doctor_id, diagnosis, prescription, notes, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW())`,
			apptID, patientID, doctorID,
			diagnosis, prescription,
			fmt.Sprintf("Treatment note %d — follow up in 2 weeks.", i+1),
		)
		if err != nil {
			return count, fmt.Errorf("treatment %d: %w", i, err)
		}
		count++
	}
	return count, nil
}

func getDoctorIDs(ctx context.Context, db *pgxpool.Pool) ([]int64, error) {
	rows, err := db.Query(ctx, `SELECT id FROM users WHERE role='doctor' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func mustRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("random: %v", err)
	}
	return hex.EncodeToString(b)
}

func randN(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

// ── static data ──────────────────────────────────────────────────────────────

var registrarFirstNames = []string{"Alice", "Bob", "Carol", "David", "Eve"}
var doctorFirstNames = []string{
	"James", "Maria", "Robert", "Linda", "Michael", "Patricia",
	"William", "Barbara", "David", "Susan", "Richard", "Jessica",
	"Joseph", "Sarah", "Thomas", "Karen", "Charles", "Nancy", "Christopher",
}
var doctorLastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones",
	"Garcia", "Miller", "Davis", "Wilson", "Martinez",
	"Anderson", "Taylor", "Thomas", "Moore", "Jackson",
	"White", "Harris", "Martin", "Thompson", "Young",
}

var firstNames = []string{
	"Emma", "Liam", "Olivia", "Noah", "Ava", "William", "Sophia", "James",
	"Isabella", "Oliver", "Mia", "Benjamin", "Charlotte", "Elijah", "Amelia",
	"Lucas", "Harper", "Mason", "Evelyn", "Logan", "Abigail", "Ethan",
	"Emily", "Aiden", "Elizabeth", "Jackson", "Sofia", "Sebastian", "Avery",
	"Mateo", "Ella", "Jack", "Scarlett", "Owen", "Grace", "Samuel", "Chloe",
	"Henry", "Victoria", "Alexander",
}

var lastNames = []string{
	"Anderson", "Bell", "Carter", "Davis", "Evans",
	"Foster", "Green", "Hall", "Irving", "Jones",
	"King", "Lee", "Morgan", "Nash", "Owen",
	"Parker", "Quinn", "Reed", "Scott", "Turner",
	"Underwood", "Vance", "Walker", "Xavier", "Young",
	"Zhang", "Adams", "Brooks", "Clarke", "Dean",
	"Ellis", "Fisher", "Gray", "Hayes", "Ingram",
	"Jensen", "Klein", "Lambert", "Mason", "Norton",
}

var medicalInfos = []string{
	"No known allergies",
	"Penicillin allergy",
	"Type 2 Diabetes — controlled with metformin",
	"Hypertension — on lisinopril",
	"Asthma — uses albuterol inhaler",
	"No significant medical history",
	"Previous appendectomy (2018)",
	"Hypothyroidism — on levothyroxine",
	"Depression — managed with therapy",
	"Seasonal allergies — antihistamines as needed",
}

var appointmentReasons = []string{
	"Annual physical examination",
	"Follow-up for hypertension management",
	"Flu symptoms — fever and cough",
	"Lower back pain assessment",
	"Routine diabetes check",
	"Skin rash evaluation",
	"Headache and dizziness",
	"Pre-operative consultation",
	"Post-operative follow-up",
	"Chronic fatigue evaluation",
	"Blood pressure monitoring",
	"Chest pain assessment",
	"Knee pain and mobility issues",
	"Anxiety and stress management",
	"Vaccination update",
}

var diagnoses = []string{
	"Hypertension — blood pressure 145/92 mmHg",
	"Type 2 Diabetes Mellitus — HbA1c 7.8%",
	"Upper respiratory tract infection",
	"Lumbar strain — muscle spasm L4-L5",
	"Gastroesophageal reflux disease (GERD)",
	"Tension headache",
	"Seasonal allergic rhinitis",
	"Anxiety disorder — generalized",
	"Hyperlipidemia — elevated LDL",
	"Vitamin D deficiency",
	"Iron deficiency anemia",
	"Osteoarthritis — right knee",
	"Migraine without aura",
	"Urinary tract infection",
	"Contact dermatitis",
	"Insomnia — chronic",
	"Hypothyroidism — under-treated",
	"Obesity — BMI 32.4",
	"Plantar fasciitis",
	"Acute sinusitis",
}

var prescriptions = []string{
	"Lisinopril 10mg once daily; low-sodium diet; follow up in 4 weeks",
	"Metformin 500mg twice daily with meals; diet and exercise counselling",
	"Amoxicillin 500mg three times daily for 7 days; rest and fluids",
	"Ibuprofen 400mg as needed; physiotherapy referral; heat application",
	"Omeprazole 20mg once daily before breakfast; avoid spicy foods",
	"Paracetamol 500mg as needed; rest; hydration",
	"Cetirizine 10mg once daily; avoid known allergens",
	"CBT referral; breathing exercises; follow up in 2 weeks",
	"Atorvastatin 20mg once daily at night; repeat lipid panel in 3 months",
	"Vitamin D3 2000 IU daily for 3 months; recheck levels",
	"Ferrous sulfate 325mg twice daily; vitamin C with each dose",
	"Naproxen 500mg twice daily; knee brace; physiotherapy",
	"Sumatriptan 50mg at onset; avoid triggers; headache diary",
	"Trimethoprim 200mg twice daily for 3 days; increase fluid intake",
	"Hydrocortisone 1% cream twice daily; avoid identified irritant",
	"Sleep hygiene education; melatonin 0.5mg if needed; CBT-I referral",
	"Increase levothyroxine to 75mcg; recheck TSH in 6 weeks",
	"Dietary counselling; exercise programme; target 5% weight loss",
	"Stretching programme; heel inserts; ice after activity",
	"Amoxicillin-clavulanate 875mg twice daily for 10 days; nasal rinse",
}
