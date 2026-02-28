import { useRef, useState } from 'react';
import { uploadTickets } from '../api/products';
import type { TicketUploadSummary } from '../types';

// ── Icons ──────────────────────────────────────────────────────────────────────

function UploadIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="17 8 12 3 7 8" />
      <line x1="12" y1="3" x2="12" y2="15" />
    </svg>
  );
}

function SpinnerIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      aria-hidden="true"
      className="uploader-spinner"
    >
      <circle cx="12" cy="12" r="9" strokeOpacity="0.25" />
      <path d="M12 3a9 9 0 0 1 9 9" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function ErrorIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="10" />
      <line x1="15" y1="9" x2="9" y2="15" />
      <line x1="9" y1="9" x2="15" y2="15" />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}

// ── Component ─────────────────────────────────────────────────────────────────

type UploadState = 'idle' | 'uploading' | 'done';

export default function TicketUploader() {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploadState, setUploadState] = useState<UploadState>('idle');
  const [summary, setSummary] = useState<TicketUploadSummary | null>(null);

  function handleButtonClick() {
    inputRef.current?.click();
  }

  async function handleFilesSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    if (files.length === 0) return;

    // Reset the input so the same files can be re-selected if needed
    e.target.value = '';

    setUploadState('uploading');
    setSummary(null);

    const result = await uploadTickets(files);

    setUploadState('done');
    setSummary(result);
  }

  function handleDismiss() {
    setUploadState('idle');
    setSummary(null);
  }

  const isUploading = uploadState === 'uploading';

  return (
    <div className="ticket-uploader">
      {/* Hidden file input — accepts multiple PDF files */}
      <input
        ref={inputRef}
        type="file"
        accept=".pdf,application/pdf"
        multiple
        aria-label="Seleccionar tickets PDF"
        className="ticket-uploader__input"
        onChange={handleFilesSelected}
      />

      {/* Upload button */}
      <button
        type="button"
        className="ticket-uploader__btn"
        onClick={handleButtonClick}
        disabled={isUploading}
        aria-label={isUploading ? 'Subiendo tickets…' : 'Subir tickets PDF'}
        title={isUploading ? 'Subiendo tickets…' : 'Subir tickets PDF'}
      >
        {isUploading ? <SpinnerIcon /> : <UploadIcon />}
        <span className="ticket-uploader__btn-label">
          {isUploading ? 'Subiendo…' : 'Subir tickets'}
        </span>
      </button>

      {/* Result toast */}
      {uploadState === 'done' && summary !== null && (
        <div
          role="status"
          aria-live="polite"
          className={`ticket-uploader__toast ${summary.failed > 0 ? 'ticket-uploader__toast--error' : 'ticket-uploader__toast--success'}`}
        >
          <div className="ticket-uploader__toast-icon">
            {summary.failed > 0 ? <ErrorIcon /> : <CheckIcon />}
          </div>
          <div className="ticket-uploader__toast-body">
            <p className="ticket-uploader__toast-title">
              {summary.failed === 0
                ? `${summary.succeeded} ticket${summary.succeeded !== 1 ? 's' : ''} importado${summary.succeeded !== 1 ? 's' : ''}`
                : `${summary.succeeded} ok · ${summary.failed} error${summary.failed !== 1 ? 'es' : ''}`}
            </p>
            {summary.items.length > 1 && (
              <ul className="ticket-uploader__toast-list">
                {summary.items.map((item) => (
                  <li
                    key={item.file}
                    className={`ticket-uploader__toast-item ${item.ok ? 'ticket-uploader__toast-item--ok' : 'ticket-uploader__toast-item--err'}`}
                  >
                    <span className="ticket-uploader__toast-filename">{item.file}</span>
                    {item.ok
                      ? ` · ${item.result.linesImported} línea${item.result.linesImported !== 1 ? 's' : ''}`
                      : ` · ${item.error?.includes('already imported') ? 'Ya importado' : item.error}`}
                  </li>
                ))}
              </ul>
            )}
          </div>
          <button
            type="button"
            className="ticket-uploader__toast-close"
            onClick={handleDismiss}
            aria-label="Cerrar notificación"
          >
            <CloseIcon />
          </button>
        </div>
      )}
    </div>
  );
}
