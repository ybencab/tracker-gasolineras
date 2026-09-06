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
      <div class="filtros">
        <select
          value={provinciaId}
          onChange={(e) => setProvinciaId((e.target as HTMLSelectElement).value)}
        >
          <option value="">Provincia...</option>
          {provincias.map((p) => (
            <option value={p.id} key={p.id}>
              {p.nombre}
            </option>
          ))}
        </select>

        <select
          value={municipioId}
          disabled={!provinciaId}
          onChange={(e) => setMunicipioId((e.target as HTMLSelectElement).value)}
        >
          <option value="">Municipio...</option>
          {municipios.map((m) => (
            <option value={m.id} key={m.id}>
              {m.nombre}
            </option>
          ))}
        </select>
      </div>

      {error && <p class="error">{error}</p>}
      {cargando && <p>Cargando estaciones...</p>}

      {estaciones.length > 0 && (
        <>
          {fecha && <p class="fecha">Precios actualizados: {fecha}</p>}
          <ul class="estaciones">
            {estaciones.map((estacion) => (
              <li key={estacion.id}>
                <strong>{estacion.rotulo}</strong> — {estacion.direccion}
                <ul class="precios">
                  {Object.entries(estacion.precios).map(([combustible, precio]) => (
                    <li key={combustible}>
                      {combustible}: {precio.toFixed(3)} €
                    </li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        </>
      )}

      {!cargando && municipioId && estaciones.length === 0 && !error && (
        <p>No hay estaciones para ese municipio.</p>
      )}
    </div>
  );
}
