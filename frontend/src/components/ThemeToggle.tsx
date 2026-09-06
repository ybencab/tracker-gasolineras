import { useEffect, useState } from "preact/hooks";

export default function ThemeToggle() {
  const [oscuro, setOscuro] = useState(false);

  useEffect(() => {
    setOscuro(document.documentElement.classList.contains("dark"));
  }, []);

  const alternar = () => {
    const nuevoValor = !oscuro;
    setOscuro(nuevoValor);
    document.documentElement.classList.toggle("dark", nuevoValor);
    localStorage.setItem("theme", nuevoValor ? "dark" : "light");
  };

  return (
    <button
      type="button"
      onClick={alternar}
      aria-label={oscuro ? "Cambiar a modo claro" : "Cambiar a modo oscuro"}
      class="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm shadow-sm transition hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-800 dark:hover:bg-slate-700"
    >
      {oscuro ? "☀️" : "🌙"}
    </button>
  );
}
