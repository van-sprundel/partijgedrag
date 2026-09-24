window.compassHistory = (() => {
  const STORAGE_KEY = "stemwijzer-history-v1";
  const LIMIT = 5;

  const read = () => {
    try {
      const list = JSON.parse(localStorage.getItem(STORAGE_KEY) || "[]");
      return Array.isArray(list) ? list.filter((entry) => entry && typeof entry.key === "string") : [];
    } catch {
      return [];
    }
  };

  const write = (list) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(list.slice(0, LIMIT)));
    } catch { /* private mode or full storage: go without history */ }
  };

  return {
    list: read,

    remember(entry) {
      write([entry, ...read().filter((existing) => existing.key !== entry.key)]);
    },

    // Fills in details only the results page knows. Unknown keys are ignored,
    // so viewing a shared link never adds it to the list.
    update(key, details) {
      const list = read();
      const entry = list.find((existing) => existing.key === key);
      if (!entry) return;
      Object.assign(entry, details);
      write(list);
    },

    clear() {
      try { localStorage.removeItem(STORAGE_KEY); } catch { /* nothing to clear */ }
    }
  };
})();
