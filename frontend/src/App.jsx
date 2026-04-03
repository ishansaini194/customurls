import { useState } from "react";
import "./App.css";
import { motion, AnimatePresence } from "framer-motion";

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

const Home = () => {
  const [url, setUrl] = useState("");
  const [alias, setAlias] = useState("");
  const [expiry, setExpiry] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setShortUrl("");

    if (!url) {
      setError("URL is required");
      return;
    }

    try {
      new URL(url);
    } catch {
      setError("Please enter a valid URL");
      return;
    }

    setLoading(true);

    try {
      const body = { url };
      if (alias) body.alias = alias;
      if (expiry) body.expiry = parseInt(expiry) * 24; // convert days to hours

      const res = await fetch(`${BACKEND_URL}/shorten`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error || "Something went wrong");
        return;
      }

      setShortUrl(data.short);
    } catch (err) {
      setError("Failed to reach the server. Is the backend running?");
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shortUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div className="min-h-screen bg-white">
      <div className="w-full max-w-7xl mx-auto px-6 sm:px-12 md:px-24 py-12 md:py-24">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-12 md:gap-24 items-center min-h-[80vh]">
          {/* Left Column - Form */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="col-span-1 flex flex-col justify-center"
          >
            <h1
              data-testid="page-title"
              className="font-display text-5xl sm:text-7xl tracking-tighter uppercase font-black text-black mb-12"
            >
              THE LINK<br />SHORTENER
            </h1>

            <form onSubmit={handleSubmit} className="flex flex-col gap-12">
              {/* URL Input */}
              <div>
                <label className="block text-sm sm:text-base uppercase tracking-[0.2em] font-bold text-black mb-4">
                  ORIGINAL URL
                </label>
                <input
                  data-testid="url-input"
                  type="text"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com/very/long/path"
                  className={`brutalist-input text-2xl sm:text-4xl ${error && !url ? "error" : ""}`}
                />
                {error && !url && <p className="error-text">{error}</p>}
                {error && url && error.includes("valid") && <p className="error-text">{error}</p>}
              </div>

              {/* Alias Input */}
              <div>
                <label className="block text-sm sm:text-base uppercase tracking-[0.2em] font-bold text-black mb-4">
                  CUSTOM ALIAS (OPTIONAL)
                </label>
                <input
                  data-testid="alias-input"
                  type="text"
                  value={alias}
                  onChange={(e) => setAlias(e.target.value)}
                  placeholder="my-cool-link"
                  className="brutalist-input text-2xl sm:text-4xl"
                />
              </div>

              {/* Expiry Input */}
              <div>
                <label className="block text-sm sm:text-base uppercase tracking-[0.2em] font-bold text-black mb-4">
                  EXPIRY IN DAYS (OPTIONAL)
                </label>
                <input
                  data-testid="expiry-input"
                  type="number"
                  value={expiry}
                  onChange={(e) => setExpiry(e.target.value)}
                  placeholder="30"
                  min="1"
                  className="brutalist-input text-2xl sm:text-4xl"
                />
              </div>

              {/* Submit Button */}
              <motion.button
                data-testid="shorten-submit-button"
                type="submit"
                disabled={loading}
                className="brutalist-button w-full sm:w-auto"
                whileTap={{ scale: 0.98 }}
              >
                {loading ? "SHORTENING..." : "SHORTEN LINK"}
              </motion.button>
            </form>

            {/* Error from server */}
            {error && url && !error.includes("valid") && (
              <p className="error-text mt-4">{error}</p>
            )}

            {/* Result */}
            <AnimatePresence>
              {shortUrl && (
                <motion.div
                  data-testid="short-url-result"
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -20 }}
                  transition={{ duration: 0.4 }}
                  className="result-card mt-16 p-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-6"
                >
                  <div>
                    <p className="text-sm uppercase tracking-[0.2em] font-bold text-black mb-2">
                      YOUR SHORT LINK:
                    </p>
                    <p
                      data-testid="shortened-url-text"
                      className="text-2xl sm:text-3xl font-bold font-mono tracking-tighter text-black break-all"
                    >
                      {shortUrl}
                    </p>
                  </div>
                  <button
                    data-testid="copy-url-button"
                    onClick={handleCopy}
                    className="copy-button whitespace-nowrap"
                  >
                    {copied ? "COPIED!" : "COPY"}
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>

          {/* Right Column - Image */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="col-span-1 hidden md:flex items-center justify-center"
          >
            <img
              src="https://images.unsplash.com/photo-1697981313183-8d309d26698c?crop=entropy&cs=srgb&fm=jpg&ixid=M3w4NjA1NzB8MHwxfHNlYXJjaHwxfHxhYnN0cmFjdCUyMGdlb21ldHJpYyUyMGFydHxlbnwwfHx8YmxhY2tfYW5kX3doaXRlfDE3NzQ5ODI0NTZ8MA&ixlib=rb-4.1.0&q=85"
              alt="Abstract geometric art"
              className="hero-image w-full max-w-md object-cover aspect-square"
            />
          </motion.div>
        </div>
      </div>
    </div>
  );
};

function App() {
  return <Home />;
}

export default App;