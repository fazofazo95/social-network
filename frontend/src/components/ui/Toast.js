"use client";

import { createContext, useContext, useState, useCallback, useRef } from "react";

const ToastContext = createContext(null);

const TOAST_TYPES = {
  success: {
    bg: "bg-green-900/80 border-green-500/60",
    icon: "✓",
    iconBg: "bg-green-500/30 text-green-300",
    bar: "bg-green-400",
  },
  error: {
    bg: "bg-red-900/80 border-red-500/60",
    icon: "✕",
    iconBg: "bg-red-500/30 text-red-300",
    bar: "bg-red-400",
  },
  warning: {
    bg: "bg-yellow-900/80 border-yellow-500/60",
    icon: "⚠",
    iconBg: "bg-yellow-500/30 text-yellow-300",
    bar: "bg-yellow-400",
  },
  info: {
    bg: "bg-purple-900/80 border-purple-500/60",
    icon: "ℹ",
    iconBg: "bg-purple-500/30 text-purple-300",
    bar: "bg-purple-400",
  },
};

function ToastItem({ toast, onDismiss }) {
  const style = TOAST_TYPES[toast.type] || TOAST_TYPES.info;

  return (
    <div
      className={`relative flex items-start gap-3 px-4 py-3 rounded-lg border backdrop-blur-md shadow-lg 
        ${style.bg} text-white min-w-75 max-w-105 animate-[slideIn_0.3s_ease-out]`}
    >
      <span className={`flex items-center justify-center w-6 h-6 rounded-full text-sm font-bold shrink-0 mt-0.5 ${style.iconBg}`}>
        {style.icon}
      </span>
      <p className="text-sm flex-1 wrap-break-word">{toast.message}</p>
      {toast.action && (
        <button
          onClick={() => {
            toast.action.onClick();
            onDismiss(toast.id);
          }}
          className="text-purple-300 hover:text-purple-100 text-sm font-semibold shrink-0 underline cursor-pointer"
        >
          {toast.action.label}
        </button>
      )}
      <button
        onClick={() => onDismiss(toast.id)}
        className="text-white/50 hover:text-white text-lg leading-none shrink-0 cursor-pointer"
      >
        ×
      </button>
      <div className={`absolute bottom-0 left-0 h-0.5 ${style.bar} rounded-b-lg animate-[shrink_${toast.duration ? toast.duration / 1000 : 3}s_linear_forwards]`} />
    </div>
  );
}

function ConfirmToast({ toast, onDismiss }) {
  return (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-9999">
      <div className="bg-[#1a1a2e] border border-purple-500/50 rounded-lg p-6 max-w-sm w-full mx-4 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
        <p className="text-white text-sm mb-6">{toast.message}</p>
        <div className="flex justify-end gap-3">
          <button
            onClick={() => {
              toast.onCancel?.();
              onDismiss(toast.id);
            }}
            className="px-4 py-2 rounded-md bg-gray-700 hover:bg-gray-600 text-white text-sm transition cursor-pointer"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              toast.onConfirm?.();
              onDismiss(toast.id);
            }}
            className="px-4 py-2 rounded-md bg-purple-600 hover:bg-purple-500 text-white text-sm transition cursor-pointer shadow-[0_0_10px_rgba(168,85,247,0.3)]"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );
}

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);
  const idRef = useRef(0);

  const dismiss = useCallback((id) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback((message, type = "info", options = {}) => {
    const id = ++idRef.current;
    const duration = options.duration || 3000;
    setToasts((prev) => [...prev, { id, message, type, isConfirm: false, action: options.action, duration }]);
    setTimeout(() => dismiss(id), duration);
  }, [dismiss]);

  const toast = useCallback(
    Object.assign((message, options) => addToast(message, "info", options), {
      success: (message, options) => addToast(message, "success", options),
      error: (message, options) => addToast(message, "error", options),
      warning: (message, options) => addToast(message, "warning", options),
      info: (message, options) => addToast(message, "info", options),
      confirm: (message, onConfirm, onCancel) => {
        const id = ++idRef.current;
        setToasts((prev) => [
          ...prev,
          { id, message, type: "info", isConfirm: true, onConfirm, onCancel },
        ]);
      },
    }),
    [addToast]
  );

  // We need a stable object, so memoize with useCallback approach
  const toastRef = useRef(toast);
  toastRef.current = toast;

  const stableToast = useCallback((...args) => toastRef.current(...args), []);
  stableToast.success = (...args) => toastRef.current.success(...args);
  stableToast.error = (...args) => toastRef.current.error(...args);
  stableToast.warning = (...args) => toastRef.current.warning(...args);
  stableToast.info = (...args) => toastRef.current.info(...args);
  stableToast.confirm = (...args) => toastRef.current.confirm(...args);

  return (
    <ToastContext.Provider value={stableToast}>
      {children}
      {/* Confirm dialogs */}
      {toasts
        .filter((t) => t.isConfirm)
        .map((t) => (
          <ConfirmToast key={t.id} toast={t} onDismiss={dismiss} />
        ))}
      {/* Toast stack */}
      <div className="fixed top-4 right-4 z-9998 flex flex-col gap-2">
        {toasts
          .filter((t) => !t.isConfirm)
          .map((t) => (
            <ToastItem key={t.id} toast={t} onDismiss={dismiss} />
          ))}
      </div>
      <style dangerouslySetInnerHTML={{ __html: `
        @keyframes slideIn {
          from { opacity: 0; transform: translateX(100%); }
          to { opacity: 1; transform: translateX(0); }
        }
        @keyframes shrink {
          from { width: 100%; }
          to { width: 0%; }
        }
      `}} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within a ToastProvider");
  return ctx;
}
