import { useState, useEffect, useRef, useCallback } from 'react';

/*
  Props:
    label        – field label string
    placeholder  – input placeholder
    onSearch     – async (query: string) => [{id, label}]
    onSelect     – (id, label) => void
    displayValue – currently selected name to show in input
    required     – boolean
*/
export default function SearchableDropdown({
  label,
  placeholder = 'Type to search…',
  onSearch,
  onSelect,
  displayValue = '',
  required = false,
}) {
  const [query, setQuery]         = useState(displayValue);
  const [results, setResults]     = useState([]);
  const [open, setOpen]           = useState(false);
  const [loading, setLoading]     = useState(false);
  const [activeIdx, setActiveIdx] = useState(-1);
  const [selected, setSelected]   = useState(!!displayValue);

  const containerRef = useRef(null);
  const debounceRef  = useRef(null);

  // Sync external displayValue changes (e.g. form reset)
  useEffect(() => {
    setQuery(displayValue);
    setSelected(!!displayValue);
  }, [displayValue]);

  // Close on outside click
  useEffect(() => {
    function handleClickOutside(e) {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const runSearch = useCallback(async (q) => {
    if (q.trim().length < 2) {
      setResults([]);
      setOpen(false);
      return;
    }
    setLoading(true);
    setOpen(true);
    try {
      const items = await onSearch(q);
      setResults(items ?? []);
      setActiveIdx(-1);
    } finally {
      setLoading(false);
    }
  }, [onSearch]);

  function handleChange(e) {
    const val = e.target.value;
    setQuery(val);
    setSelected(false);
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => runSearch(val), 300);
  }

  function handleSelect(item) {
    setQuery(item.label);
    setSelected(true);
    setOpen(false);
    setResults([]);
    onSelect(item.id, item.label);
  }

  function handleKeyDown(e) {
    if (!open) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setActiveIdx((i) => Math.min(i + 1, results.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setActiveIdx((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (activeIdx >= 0 && results[activeIdx]) handleSelect(results[activeIdx]);
    } else if (e.key === 'Escape') {
      setOpen(false);
    }
  }

  return (
    <div className="form-group" ref={containerRef}>
      <label>{label}{required && ' *'}</label>
      <div className="sdd-wrap">
        <input
          type="text"
          value={query}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          onFocus={() => { if (results.length > 0) setOpen(true); }}
          placeholder={placeholder}
          autoComplete="off"
          className={selected ? 'sdd-input sdd-selected' : 'sdd-input'}
          required={required && !selected}
        />
        {/* hidden input carries the real ID for form validation */}
        {required && <input type="hidden" value={selected ? '1' : ''} required />}

        {open && (
          <ul className="sdd-list" role="listbox">
            {loading && (
              <li className="sdd-item sdd-loading">
                <span className="sdd-spinner" /> Searching…
              </li>
            )}
            {!loading && results.length === 0 && (
              <li className="sdd-item sdd-empty">No results found</li>
            )}
            {!loading && results.map((item, idx) => (
              <li
                key={item.id}
                className={`sdd-item${idx === activeIdx ? ' sdd-active' : ''}`}
                role="option"
                aria-selected={idx === activeIdx}
                onMouseDown={(e) => { e.preventDefault(); handleSelect(item); }}
                onMouseEnter={() => setActiveIdx(idx)}
              >
                {item.label}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
