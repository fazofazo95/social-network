import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen bg-(--color-customPurple) flex flex-col items-center justify-center px-4">
      <h1 className="text-[8rem] font-black text-purple-500 leading-none neon-glow select-none">
        404
      </h1>
      <p className="text-2xl font-semibold text-purple-200 mt-2 mb-1">
        Page Not Found
      </p>
      <p className="text-purple-400 text-center max-w-md mb-8">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <Link
        href="/"
        className="px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white font-semibold rounded-lg transition shadow-custom"
      >
        Back to Home
      </Link>
    </div>
  );
}
