// Cliente API para comunicar con el backend en Go y PostgreSQL (Railway)
const API_BASE = '/api';

const toCamel = (key) => key.replace(/_([a-z])/g, (_, c) => c.toUpperCase());

const normalize = (value) => {
    if (Array.isArray(value)) return value.map(normalize);
    if (!value || typeof value !== 'object') return value;
    return Object.fromEntries(
        Object.entries(value).map(([key, item]) => [toCamel(key), normalize(item)])
    );
};

export async function obtenerPedidos() {
    const response = await fetch(`${API_BASE}/pedidos`);
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Error del servidor (${response.status})`);
    }
    const data = await response.json();
    return normalize(data).map(item => ({
        id: item.id || item.pedidoId || item.idPedido,
        ...item
    }));
}

export async function crearPedido(data) {
    const response = await fetch(`${API_BASE}/pedidos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Error creando pedido (${response.status})`);
    }
    const result = await response.json();
    const normalized = normalize(result);
    return {
        id: normalized.id || normalized.pedidoId || normalized.idPedido,
        ...normalized
    };
}

export async function actualizarEstadoPedido(id, nuevoEstado) {
    const response = await fetch(`${API_BASE}/pedidos/${encodeURIComponent(id)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ estado: nuevoEstado })
    });
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Error actualizando estado (${response.status})`);
    }
    return response.json();
}

export async function eliminarPedido(id) {
    const response = await fetch(`${API_BASE}/pedidos/${encodeURIComponent(id)}`, {
        method: 'DELETE'
    });
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Error eliminando pedido (${response.status})`);
    }
    return true;
}
