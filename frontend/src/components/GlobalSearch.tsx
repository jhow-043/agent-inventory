// Global search component — quick device lookup by hostname.

import { useState, useRef, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { getDevices } from '../api/devices';
import { useDebounce } from '../hooks/useDebounce';

export default function GlobalSearch() {
  const [query, setQuery] = useState('');
  const [open, setOpen] = useState(false);
  const debouncedQuery = useDebounce(query, 250);
  const navigate = useNavigate();
  const ref = useRef<HTMLDivElement>(null);

  const { data } = useQuery({
    queryKey: ['search', debouncedQuery],
    queryFn: () => getDevices({ hostname: debouncedQuery, limit: 8 }),
    enabled: debouncedQuery.length >= 2,
  });

  const results = data?.devices ?? [];

  // Close on click outside
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const handleSelect = (hostname: string) => {
    navigate(`/devices/${encodeURIComponent(hostname)}`);
    setQuery('');
    setOpen(false);
  };

  return (
    <div ref={ref} className="relative w-full max-w-md">
      <div className="relative">
        <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
        </svg>
        <input
          type="text"
          value={query}
          onChange={(e) => { setQuery(e.target.value); setOpen(true); }}
          onFocus={() => setOpen(true)}
          placeholder="Buscar dispositivos..."
          className="w-full pl-9 pr-4 py-2 text-sm rounded-lg border border-border-primary bg-bg-tertiary/50 text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-accent/40 focus:border-accent/40 transition-all"
        />
        {query && (
          <button
            onClick={() => { setQuery(''); setOpen(false); }}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary transition-colors"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>

      {/* Results dropdown */}
      {open && debouncedQuery.length >= 2 && (
        <div className="absolute top-full left-0 right-0 mt-1 bg-bg-secondary border border-border-primary rounded-xl shadow-xl z-50 overflow-hidden animate-scale-in">
          {results.length === 0 ? (
            <div className="px-4 py-3 text-sm text-text-muted">Nenhum dispositivo encontrado</div>
          ) : (
            <ul>
              {results.map((d) => (
                <li key={d.id}>
                  <button
                    onClick={() => handleSelect(d.hostname)}
                    className="w-full flex items-center gap-3 px-4 py-2.5 text-left hover:bg-bg-tertiary transition-colors cursor-pointer"
                  >
                    <svg className="w-4 h-4 text-text-muted flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25A2.25 2.25 0 015.25 3h13.5A2.25 2.25 0 0121 5.25z" />
                    </svg>
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium text-text-primary truncate">{d.hostname}</div>
                      <div className="text-xs text-text-muted truncate">{d.os_name} • {d.serial_number}</div>
                    </div>
                    <span className={`w-2 h-2 rounded-full flex-shrink-0 ${d.status === 'active' ? 'bg-success' : 'bg-warning'}`} />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
