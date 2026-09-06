import { useEffect, useMemo, useState } from "preact/hooks";

const API_URL = import.meta.env.PUBLIC_API_URL;

interface Provincia {
  id: string;
  nombre: string;
}

interface Municipio {
  id: string;
  nombre: string;
}

interface Estacion {
  id: string;
  rotulo: string;
  direccion: string;
  localidad: string;
  precios: Record<string, number>;
}

interface EstacionesResponse {
  fecha: string;
  estaciones: Estacion[];
}

function formatPrecio(precio: number): string {
  return precio.toLocaleString("es-ES", { minimumFractionDigits: 3, maximumFractionDigits: 3 });
}

const selectClass =
  "w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-700 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 dark:focus:ring-sky-900 dark:disabled:bg-slate-900 dark:disabled:text-slate-500";

export default function Buscador() {
  const [provincias, setProvincias] = useState<Provincia[]>([]);
  const [municipios, setMunicipios] = useState<Municipio[]>([]);
  const [estaciones, setEstaciones] = useState<Estacion[]>([]);
  const [fecha, setFecha] = useState<string | null>(null);
  const [provinciaId, setProvinciaId] = useState("");
  const [municipioId, setMunicipioId] = useState("");
  const [combustible, setCombustible] = useState("");
  const [ordenAscendente, setOrdenAscendente] = useState(true);
  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch(`${API_URL}/api/provincias`)
      .then((r) => r.json())
      .then((data: Provincia[]) =>
        setProvincias([...data].sort((a, b) => a.nombre.localeCompare(b.nombre))),
      )
      .catch(() => setError("No se han podido cargar las provincias."));
  }, []);

  useEffect(() => {
    setMunicipios([]);
    setMunicipioId("");
    setEstaciones([]);
    if (!provinciaId) return;

    fetch(`${API_URL}/api/municipios?provincia=${provinciaId}`)
      .then((r) => r.json())
      .then((data: Municipio[]) =>
        setMunicipios([...data].sort((a, b) => a.nombre.localeCompare(b.nombre))),
      )
      .catch(() => setError("No se han podido cargar los municipios."));
  }, [provinciaId]);

  useEffect(() => {
    setEstaciones([]);
    setCombustible("");
    if (!municipioId) return;

    setCargando(true);
    setError(null);
    fetch(`${API_URL}/api/estaciones?municipio=${municipioId}`)
      .then((r) => r.json())
      .then((data: EstacionesResponse) => {
        setEstaciones(data.estaciones);
        setFecha(data.fecha);
      })
      .catch(() => setError("No se han podido cargar las estaciones."))
      .finally(() => setCargando(false));
  }, [municipioId]);

  const combustiblesDisponibles = useMemo(() => {
    const nombres = new Set<string>();
    for (const estacion of estaciones) {
      for (const nombre of Object.keys(estacion.precios)) nombres.add(nombre);
    }
    return [...nombres].sort();
  }, [estaciones]);

  const estacionesFiltradas = useMemo(() => {
    if (!combustible) return estaciones;
    const signo = ordenAscendente ? 1 : -1;
    return estaciones
      .filter((e) => combustible in e.precios)
      .sort((a, b) => signo * (a.precios[combustible] - b.precios[combustible]));
  }, [estaciones, combustible, ordenAscendente]);

  return (
    <div>
      <div class="grid grid-cols-1 gap-4 rounded-xl border border-slate-200 bg-white p-5 shadow-sm sm:grid-cols-2 dark:border-slate-700 dark:bg-slate-800">
        <label class="block">
          <span class="mb-1 block text-sm font-medium text-slate-600 dark:text-slate-300">Provincia</span>
          <select
            class={selectClass}
            value={provinciaId}
            onChange={(e) => setProvinciaId((e.target as HTMLSelectElement).value)}
          >
            <option value="">Selecciona una provincia...</option>
            {provincias.map((p) => (
              <option value={p.id} key={p.id}>
                {p.nombre}
              </option>
            ))}
          </select>
        </label>

        <label class="block">
          <span class="mb-1 block text-sm font-medium text-slate-600 dark:text-slate-300">Municipio</span>
          <select
            class={selectClass}
            value={municipioId}
            disabled={!provinciaId}
            onChange={(e) => setMunicipioId((e.target as HTMLSelectElement).value)}
          >
            <option value="">Selecciona un municipio...</option>
            {municipios.map((m) => (
              <option value={m.id} key={m.id}>
                {m.nombre}
              </option>
            ))}
          </select>
        </label>

        <label class="block sm:col-span-2">
          <span class="mb-1 block text-sm font-medium text-slate-600 dark:text-slate-300">Combustible</span>
          <select
            class={selectClass}
            value={combustible}
            disabled={combustiblesDisponibles.length === 0}
            onChange={(e) => setCombustible((e.target as HTMLSelectElement).value)}
          >
            <option value="">Todos</option>
            {combustiblesDisponibles.map((nombre) => (
              <option value={nombre} key={nombre}>
                {nombre}
              </option>
            ))}
          </select>
        </label>

        {combustible && (
          <div class="block sm:col-span-2">
            <button
              type="button"
              onClick={() => setOrdenAscendente((v) => !v)}
              class="inline-flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-700 shadow-sm transition hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700"
            >
              {ordenAscendente ? "↑ Menor a mayor precio" : "↓ Mayor a menor precio"}
            </button>
          </div>
        )}
      </div>

      {error && (
        <p class="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">
          {error}
        </p>
      )}

      {cargando && (
        <p class="mt-6 text-center text-slate-400 dark:text-slate-500">Cargando estaciones...</p>
      )}

      {estacionesFiltradas.length > 0 && (
        <div class="mt-6">
          {fecha && (
            <p class="mb-3 text-sm text-slate-400 dark:text-slate-500">
              Precios actualizados: {fecha}
            </p>
          )}
          <ul class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {estacionesFiltradas.map((estacion) => (
              <li
                key={estacion.id}
                class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:shadow-md dark:border-slate-700 dark:bg-slate-800"
              >
                <p class="font-semibold text-slate-900 dark:text-white">{estacion.rotulo}</p>
                <p class="mb-3 text-sm text-slate-500 dark:text-slate-400">{estacion.direccion}</p>
                <ul class="flex flex-wrap gap-2">
                  {Object.entries(estacion.precios).map(([nombreCombustible, precio]) => (
                    <li
                      key={nombreCombustible}
                      class={
                        nombreCombustible === combustible
                          ? "rounded-full bg-sky-600 px-3 py-1 text-xs font-semibold text-white"
                          : "rounded-full bg-sky-50 px-3 py-1 text-xs font-medium text-sky-700 dark:bg-sky-950/60 dark:text-sky-300"
                      }
                    >
                      {nombreCombustible}: {formatPrecio(precio)} €
                    </li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        </div>
      )}

      {!cargando && municipioId && estaciones.length > 0 && estacionesFiltradas.length === 0 && (
        <p class="mt-6 text-center text-slate-400 dark:text-slate-500">
          Ninguna estación de este municipio vende ese combustible.
        </p>
      )}

      {!cargando && municipioId && estaciones.length === 0 && !error && (
        <p class="mt-6 text-center text-slate-400 dark:text-slate-500">
          No hay estaciones para ese municipio.
        </p>
      )}
    </div>
  );
}
