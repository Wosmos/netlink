'use client';

import { useState } from 'react';

interface EmojiPickerProps {
  onSelect: (emoji: string) => void;
  onClose: () => void;
}

// Most commonly used emojis (lightweight, no external library)
const QUICK_EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '🙏'];

const EMOJI_CATEGORIES = {
  'Smileys': ['😀', '😃', '😄', '😁', '😅', '😂', '🤣', '😊', '😇', '🙂', '🙃', '😉', '😌', '😍', '🥰', '😘', '😗', '😙', '😚', '😋', '😛', '😝', '😜', '🤪', '🤨', '🧐', '🤓', '😎', '🥸', '🤩', '🥳'],
  'Gestures': ['👍', '👎', '👌', '✌️', '🤞', '🤟', '🤘', '🤙', '👈', '👉', '👆', '👇', '☝️', '👏', '🙌', '👐', '🤲', '🤝', '🙏'],
  'Hearts': ['❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❤️‍🔥', '❤️‍🩹', '💕', '💞', '💓', '💗', '💖', '💘', '💝'],
  'Faces': ['😐', '😑', '😶', '🙄', '😏', '😣', '😥', '😮', '🤐', '😯', '😪', '😫', '🥱', '😴', '😌', '😛', '😜', '😝', '🤤', '😒', '😓', '😔', '😕', '🙃', '🫠', '🤗', '🤔', '🫣', '🤭', '🫢'],
  'Symbols': ['✅', '❌', '⭐', '🔥', '💯', '✨', '💫', '⚡', '💥', '🎉', '🎊', '🎈', '🎁', '🏆', '🥇', '🥈', '🥉'],
};

export default function EmojiPicker({ onSelect, onClose }: EmojiPickerProps) {
  const [activeCategory, setActiveCategory] = useState<string>('Smileys');

  return (
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 z-40" onClick={onClose}></div>
      
      {/* Picker */}
      <div className="absolute bottom-full right-0 mb-2 bg-[#0c0c14] border border-cyan-500/30 rounded shadow-2xl z-50 w-80 max-w-[90vw]">
        {/* Quick reactions */}
        <div className="p-2 border-b border-cyan-900/30 flex gap-1">
          {QUICK_EMOJIS.map(emoji => (
            <button
              key={emoji}
              onClick={() => {
                onSelect(emoji);
                onClose();
              }}
              className="text-2xl hover:bg-cyan-950/30 rounded p-1 transition-colors"
            >
              {emoji}
            </button>
          ))}
        </div>

        {/* Categories */}
        <div className="flex border-b border-cyan-900/30 overflow-x-auto no-scrollbar">
          {Object.keys(EMOJI_CATEGORIES).map(category => (
            <button
              key={category}
              onClick={() => setActiveCategory(category)}
              className={`px-3 py-2 text-xs font-mono uppercase whitespace-nowrap ${
                activeCategory === category
                  ? 'text-cyan-400 border-b-2 border-cyan-500'
                  : 'text-gray-500 hover:text-cyan-400'
              }`}
            >
              {category}
            </button>
          ))}
        </div>

        {/* Emoji grid */}
        <div className="p-2 grid grid-cols-8 gap-1 max-h-48 overflow-y-auto">
          {EMOJI_CATEGORIES[activeCategory as keyof typeof EMOJI_CATEGORIES].map((emoji, idx) => (
            <button
              key={idx}
              onClick={() => {
                onSelect(emoji);
                onClose();
              }}
              className="text-2xl hover:bg-cyan-950/30 rounded p-1 transition-colors"
              title={emoji}
            >
              {emoji}
            </button>
          ))}
        </div>
      </div>
    </>
  );
}
