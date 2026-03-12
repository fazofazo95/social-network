'use client';
import { useState } from "react";
import { addReaction, removeReaction } from "src/lib/services/reaction";

// Accept postId as a prop to know which post to react to
const Ripple_Button = ({ postId, initialRippled = false, initialCount = 0, onChange }) => {
  const [rippled, setRippled] = useState(initialRippled);
  const [count, setCount] = useState(initialCount);
  const [loading, setLoading] = useState(false);

  const handleRipple = async () => {
    if (!postId || loading) return;
    setLoading(true);
    try {
      if (rippled) {
        // Remove reaction
        const res = await removeReaction(postId);
        setRippled(false);
        setCount(res?.like_count ?? count - 1);
        if (onChange) onChange(res?.like_count ?? count - 1, false);
      } else {
        // Add reaction
        const res = await addReaction(postId);
        setRippled(true);
        setCount(res?.like_count ?? count + 1);
        if (onChange) onChange(res?.like_count ?? count + 1, true);
      }
    } catch (e) {
      // Optionally show error
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      id="ripple_btn"
      className="flex justify-center cursor-pointer gap-1"
      onClick={handleRipple}
      disabled={loading}
      aria-pressed={rippled}
    >
      <span className="text-lg">{rippled ? "💜" : "🤍"}</span>
      Ripple
    </button>
  );
};

export default Ripple_Button;