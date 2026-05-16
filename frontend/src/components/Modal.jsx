import { useEffect, useRef } from 'react';

export default function Modal({ title, onClose, children }) {
  const modalRef   = useRef(null);
  // Store onClose in a ref so the keyboard/backdrop handlers always call the
  // latest version without being listed as effect dependencies.
  // This is the critical fix: if onClose were in the useEffect dep array,
  // every parent re-render (e.g. every keystroke in a form input) would
  // recreate the onClose arrow function, trigger the effect cleanup+re-run,
  // and call modalRef.current?.focus() — stealing focus from the active input.
  const onCloseRef = useRef(onClose);
  useEffect(() => { onCloseRef.current = onClose; });

  useEffect(() => {
    // Capture element that was focused before the modal opened so we can
    // restore it on close.
    const prev = document.activeElement;

    // Focus the modal container once on mount so keyboard users can navigate.
    // tabIndex={-1} means it receives focus programmatically but is NOT in the
    // tab order, so pressing Tab immediately goes to the first form field.
    modalRef.current?.focus();

    function onKey(e) {
      if (e.key === 'Escape') {
        e.stopPropagation();
        onCloseRef.current();
        return;
      }
      // Focus trap: keep Tab navigation inside the modal
      if (e.key === 'Tab' && modalRef.current) {
        const focusable = Array.from(modalRef.current.querySelectorAll(
          'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
        ));
        if (focusable.length === 0) return;
        const first = focusable[0];
        const last  = focusable[focusable.length - 1];
        if (e.shiftKey) {
          if (document.activeElement === first) { e.preventDefault(); last.focus(); }
        } else {
          if (document.activeElement === last)  { e.preventDefault(); first.focus(); }
        }
      }
    }

    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('keydown', onKey);
      prev?.focus(); // restore focus to the element that opened the modal
    };
  }, []); // ← empty: runs ONCE on mount only, never on re-renders

  return (
    <div
      className="modal-overlay"
      role="dialog"
      aria-modal="true"
      onClick={() => onCloseRef.current()}
    >
      <div
        className="modal"
        ref={modalRef}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="modal-header">
          <h3 className="modal-title">{title}</h3>
          <button
            className="modal-close"
            onClick={() => onCloseRef.current()}
            aria-label="Close"
          >
            &times;
          </button>
        </div>
        <div className="modal-body">{children}</div>
      </div>
    </div>
  );
}
