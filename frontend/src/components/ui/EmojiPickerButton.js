"use client";

import { useState, useRef, useEffect } from "react";
import EmojiPicker from "emoji-picker-react";

const EmojiPickerButton = ({ onEmojiSelect, className = "" }) => {
  const [open, setOpen] = useState(false);
  const pickerRef = useRef(null);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (pickerRef.current && !pickerRef.current.contains(e.target)) {
        setOpen(false);
      }
    };
    if (open) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [open]);

  return (
    <div className={`relative ${className}`} ref={pickerRef}>
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        className="text-lg cursor-pointer hover:scale-110 transition-transform select-none"
        title="Add emoji"
      >
        😀
      </button>
      {open && (
        <div className="absolute top-8 right-0 z-50">
          <EmojiPicker
            onEmojiClick={(emojiData) => {
              onEmojiSelect(emojiData.emoji);
              setOpen(false);
            }}
            width={300}
            height={380}
            searchDisabled={false}
            skinTonesDisabled
            previewConfig={{ showPreview: false }}
          />
        </div>
      )}
    </div>
  );
};

export default EmojiPickerButton;
