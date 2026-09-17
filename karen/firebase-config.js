// Compatibilidad con la interfaz que usaba la aplicación. Las operaciones
// ahora pasan por el backend Go y nunca exponen las credenciales de PostgreSQL.
const api = '/api';
export const db = {};
export const collection = (_db, name) => ({ name });
export const orderBy = () => null;
export const query = (ref) => ref;
export const doc = (_db, collectionName, id) => ({ collection: collectionName, id });
export async function getDocs(ref) {
    const response = await fetch(`${api}/${ref.name}`);
    if (!response.ok) throw new Error(await response.text());
    const rows = await response.json();
    return { docs: rows.map(row => ({ id: row.id, data: () => row })) };
}
export async function addDoc(ref, data) {
    const response = await fetch(`${api}/${ref.name}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) });
    if (!response.ok) throw new Error(await response.text());
    const row = await response.json();
    return { id: row.id, data: () => row };
}
export async function updateDoc(reference, data) {
    const response = await fetch(`${api}/${reference.collection}/${encodeURIComponent(reference.id)}`, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) });
    if (!response.ok) throw new Error(await response.text());
}
export async function deleteDoc(reference) {
    const response = await fetch(`${api}/${reference.collection}/${encodeURIComponent(reference.id)}`, { method: 'DELETE' });
    if (!response.ok) throw new Error(await response.text());
}
