import { useEffect, useState } from "preact/hooks";

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

const selectClass =
  "w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-700 shadow-sm focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400";

export default function Buscador() {
  const [provincias, setProvincias] = useState<Provincia[]>([]);
  const [municipios, setMunicipios] = useState<Municipio[]>([]);
  const [estaciones, setEstaciones] = useState<Estacion[]>([]);
  const [fecha, setFecha] = useState<string | null>(null);
  const [provinciaId, setProvinciaId] = useState("");
  const [municipioId, setMunicipioId] = useState("");
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

  return (
    <div>
      <div class="grid grid-cols-1 gap-4 rounded-xl border border-slate-200 bg-white p-5 shadow-sm sm:grid-cols-2">
        <label class="block">
          <span class="mb-1 block text-sm font-medium text-slate-600">Provincia</span>
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
          <span class="mb-1 block text-sm font-medium text-slate-600">Municipio</span>
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
      </div>

      {error && (
        <p class="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</p>
      )}

      {cargando && <p class="mt-6 text-center text-slate-400">Cargando estaciones...</p>}

      {estaciones.length > 0 && (
        <div class="mt-6">
          {fecha && (
            <p class="mb-3 text-sm text-slate-400">Precios actualizados: {fecha}</p>
          )}
          <ul class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {estaciones.map((estacion) => (
              <li
                key={estacion.id}
                class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:shadow-md"
              >
                <p class="font-semibold text-slate-900">{estacion.rotulo}</p>
                <p class="mb-3 text-sm text-slate-500">{estacion.direccion}</p>
                <ul class="flex flex-wrap gap-2">
                  {Object.entries(estacion.precios).map(([combustible, precio]) => (
                    <li
                      key={combustible}
                      class="rounded-full bg-sky-50 px-3 py-1 text-xs font-medium text-sky-700"
                    >
                      {combustible}: {precio.toFixed(3)} €
                    </li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        </div>
      )}

      {!cargando && municipioId && estaciones.length === 0 && !error && (
        <p class="mt-6 text-center text-slate-400">No hay estaciones para ese municipio.</p>
      )}
    </div>
  );
}
