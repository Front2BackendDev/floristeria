import { obtenerPedidos, crearPedido, actualizarEstadoPedido, eliminarPedido as apiEliminarPedido } from './api.js';

const CONFIG = {
    PASSWORD_ADMIN: '2002',
    NUMERO_WHATSAPP: '573148007135',
    INTERVALO_NOTIF: 20 * 60 * 1000
};

const state = {
    pedidos: [],
    pedidoAEliminar: null,
    itemsFlores: [],
    constructorFlor: { tipo: null, colores: [] },
    gruposColapsados: new Set(),
    gruposInicializados: new Set()
};

// Colores adicionales que se ofrecen para cualquier tipo de flor (tinturados/teñidos)
const COLORES_ADICIONALES = [
    { nombre: 'Azul Celeste', color: '#7dd3fc' },
    { nombre: 'Negro', color: '#18181b' },
    { nombre: 'Blanco Hueso', color: '#e8dfc9' },
    { nombre: 'Rosado Pastel', color: '#fbcfe8' },
    { nombre: 'Lila', color: '#ddd6fe' }
];

// Catálogo de flores disponibles: cada tipo muestra primero sus colores reales/tradicionales
// y luego los colores adicionales (tinturados), para que el cliente aprenda las opciones a medida que arma su pedido.
const CATALOGO_FLORES = [
    { id: 'rosas', nombre: 'Rosas', emoji: '🌹', colores: [
        { nombre: 'Rojo', color: '#dc2626' },
        { nombre: 'Rosado', color: '#f472b6' },
        { nombre: 'Blanco', color: '#f8fafc' },
        { nombre: 'Amarillo', color: '#facc15' },
        { nombre: 'Bicolor', color: 'linear-gradient(90deg,#dc2626,#f8fafc)' },
        ...COLORES_ADICIONALES
    ]},
    { id: 'tulipanes', nombre: 'Tulipanes', emoji: '🌷', colores: [
        { nombre: 'Rojo', color: '#dc2626' },
        { nombre: 'Rosado', color: '#f472b6' },
        { nombre: 'Blanco', color: '#f8fafc' },
        { nombre: 'Amarillo', color: '#facc15' },
        { nombre: 'Morado', color: '#a855f7' },
        ...COLORES_ADICIONALES
    ]},
    { id: 'girasoles', nombre: 'Girasoles', emoji: '🌻', colores: [
        { nombre: 'Amarillo', color: '#facc15' },
        { nombre: 'Naranja', color: '#fb923c' },
        ...COLORES_ADICIONALES
    ]},
    { id: 'lirios', nombre: 'Lirios', emoji: '🌸', colores: [
        { nombre: 'Blanco', color: '#f8fafc' },
        { nombre: 'Rosado', color: '#f472b6' },
        { nombre: 'Naranja', color: '#fb923c' },
        ...COLORES_ADICIONALES
    ]},
    { id: 'orquideas', nombre: 'Orquídeas', emoji: '🪷', colores: [
        { nombre: 'Blanco', color: '#f8fafc' },
        { nombre: 'Morado', color: '#a855f7' },
        { nombre: 'Rosado', color: '#f472b6' },
        ...COLORES_ADICIONALES
    ]},
    { id: 'mixto', nombre: 'Ramo Mixto', emoji: '💐', colores: [
        { nombre: 'Multicolor', color: 'conic-gradient(from 0deg,#dc2626,#facc15,#22c55e,#3b82f6,#a855f7,#dc2626)' },
        { nombre: 'Tonos Pastel', color: 'linear-gradient(90deg,#fbcfe8,#fef9c3,#dbeafe)' },
        { nombre: 'Tonos Cálidos', color: 'linear-gradient(90deg,#fca5a5,#fdba74,#fde047)' },
        ...COLORES_ADICIONALES
    ]},
    { id: 'otras', nombre: 'Otras', emoji: '🌼', colores: [
        { nombre: 'Rojo', color: '#dc2626' },
        { nombre: 'Rosado', color: '#f472b6' },
        { nombre: 'Blanco', color: '#f8fafc' },
        { nombre: 'Amarillo', color: '#facc15' },
        { nombre: 'Naranja', color: '#fb923c' },
        { nombre: 'Morado', color: '#a855f7' },
        ...COLORES_ADICIONALES
    ]}
];

// --- INICIALIZACIÓN ---
console.log('Sistema Karen\'s Floristik Iniciado (Constructor de Flores Interactivo)');

function inicializar() {
    cargarPedidos();
    iniciarPollingInteligente();
    const anioEl = document.getElementById('anioActual');
    if (anioEl) anioEl.innerText = new Date().getFullYear();
    if (document.getElementById('grid-tipos-flor')) {
        renderGridTiposFlor();
        renderListaFlores();
    }
}

// Consulta periódica de pedidos cuando la pestaña está visible
// (evita llamadas innecesarias cuando la pestaña está en segundo plano).
let intervaloPollingId = null;

function iniciarPollingInteligente() {
    const activar = () => {
        if (intervaloPollingId) return;
        intervaloPollingId = setInterval(cargarPedidos, CONFIG.INTERVALO_NOTIF);
    };
    const detener = () => {
        if (!intervaloPollingId) return;
        clearInterval(intervaloPollingId);
        intervaloPollingId = null;
    };

    if (document.visibilityState === 'visible') activar();

    document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') {
            cargarPedidos(); // refresco inmediato al volver a la pestaña
            activar();
        } else {
            detener();
        }
    });
}

// --- CONSTRUCTOR DE FLORES ---
function renderGridTiposFlor() {
    const grid = document.getElementById('grid-tipos-flor');
    grid.innerHTML = CATALOGO_FLORES.map(f => `
        <button type="button" onclick="window.seleccionarTipoFlor('${f.id}')" data-tipo="${f.id}"
            class="tipo-flor-btn flex flex-col items-center justify-center gap-1 py-3 rounded-lg border-2 border-rose-100 bg-white hover:border-rose-300 transition-all">
            <span class="text-2xl">${f.emoji}</span>
            <span class="text-[11px] font-bold text-slate-600">${f.nombre}</span>
        </button>
    `).join('');
}

function seleccionarTipoFlor(tipoId) {
    state.constructorFlor = { tipo: tipoId, colores: [] };

    document.querySelectorAll('.tipo-flor-btn').forEach(btn => {
        btn.classList.toggle('border-rose-500', btn.dataset.tipo === tipoId);
        btn.classList.toggle('ring-2', btn.dataset.tipo === tipoId);
        btn.classList.toggle('ring-rose-200', btn.dataset.tipo === tipoId);
        btn.classList.toggle('bg-rose-50', btn.dataset.tipo === tipoId);
    });

    const flor = CATALOGO_FLORES.find(f => f.id === tipoId);
    document.getElementById('tipoFlorSeleccionadaLabel').innerText = `(${flor.nombre})`;

    const panelPersonalizada = document.getElementById('panel-flor-personalizada');
    panelPersonalizada.classList.toggle('hidden', tipoId !== 'otras');
    document.getElementById('nombreFlorPersonalizada').value = '';

    renderRuedaColores(flor);
    document.getElementById('colorPersonalizado').value = '';
    renderColoresSeleccionados();

    document.getElementById('panel-color-flor').classList.remove('hidden');
    document.getElementById('panel-cantidad-flor').classList.add('hidden');
}

// Dibuja los colores del tipo elegido como una rueda circular (en vez de una lista plana).
function renderRuedaColores(flor) {
    const cont = document.getElementById('rueda-colores-swatches');
    const colores = flor.colores;
    const total = colores.length;
    const centro = 110, radio = 86;

    cont.innerHTML = colores.map((c, i) => {
        const angulo = (360 / total) * i - 90;
        const rad = angulo * Math.PI / 180;
        const x = (centro + radio * Math.cos(rad)).toFixed(1);
        const y = (centro + radio * Math.sin(rad)).toFixed(1);
        return `
        <button type="button" onclick="window.seleccionarColorFlor('${c.nombre.replace(/'/g, "\\'")}')" data-color="${c.nombre}" title="${c.nombre}"
            class="color-flor-btn absolute w-10 h-10 rounded-full border-[3px] border-white shadow-md hover:scale-125 hover:z-20 transition-all duration-300 ease-out"
            style="left:${x}px; top:${y}px; transform:translate(-50%,-50%); background:${c.color};">
        </button>`;
    }).join('');

    document.getElementById('rueda-emoji-central').innerText = flor.emoji;
}

// Repinta el anillo de selección sobre la rueda y la lista de chips con TODOS los colores
// elegidos (los de la rueda + los escritos a mano), y muestra/oculta el paso de cantidad.
function renderColoresSeleccionados() {
    const seleccionados = state.constructorFlor.colores;

    document.querySelectorAll('.color-flor-btn').forEach(btn => {
        const activo = seleccionados.includes(btn.dataset.color);
        btn.classList.toggle('ring-4', activo);
        btn.classList.toggle('ring-rose-500', activo);
        btn.classList.toggle('ring-offset-1', activo);
        btn.classList.toggle('scale-125', activo);
        btn.classList.toggle('z-20', activo);
    });

    const cont = document.getElementById('chips-colores-seleccionados');
    cont.innerHTML = seleccionados.length
        ? seleccionados.map(c => `
            <span class="flex items-center gap-1 bg-rose-100 text-rose-700 text-xs font-bold pl-3 pr-1 py-1 rounded-full fade-in">
                ${c}
                <button type="button" onclick="window.eliminarColorPersonalizado('${c.replace(/'/g, "\\'")}')" class="hover:text-red-600 px-1.5"><i class="fas fa-times"></i></button>
            </span>
        `).join('')
        : '<span class="text-xs text-slate-400 italic">Tocá uno o más colores en la rueda 👆</span>';

    document.getElementById('panel-cantidad-flor').classList.toggle('hidden', seleccionados.length === 0);
}

function seleccionarColorFlor(colorNombre) {
    const idx = state.constructorFlor.colores.indexOf(colorNombre);
    if (idx === -1) state.constructorFlor.colores.push(colorNombre);
    else state.constructorFlor.colores.splice(idx, 1);
    renderColoresSeleccionados();
}

function agregarColorPersonalizado() {
    const input = document.getElementById('colorPersonalizado');
    const valor = input.value.trim();
    if (!valor) return;
    const yaExiste = state.constructorFlor.colores.some(c => c.toLowerCase() === valor.toLowerCase());
    if (!yaExiste) state.constructorFlor.colores.push(valor);
    input.value = '';
    renderColoresSeleccionados();
}

function eliminarColorPersonalizado(colorNombre) {
    state.constructorFlor.colores = state.constructorFlor.colores.filter(c => c !== colorNombre);
    renderColoresSeleccionados();
}

function cambiarCantidadFlor(delta) {
    const input = document.getElementById('cantidadFlorActual');
    let val = (parseInt(input.value, 10) || 1) + delta;
    if (val < 1) val = 1;
    if (val > 200) val = 200;
    input.value = val;
}

function agregarFlorAlPedido() {
    const { tipo, colores } = state.constructorFlor;
    if (!tipo || !colores.length) return;

    const flor = CATALOGO_FLORES.find(f => f.id === tipo);
    const cantidad = parseInt(document.getElementById('cantidadFlorActual').value, 10) || 1;
    const nombrePersonalizado = document.getElementById('nombreFlorPersonalizada').value.trim();
    const nombreFinal = (tipo === 'otras' && nombrePersonalizado) ? nombrePersonalizado : flor.nombre;

    state.itemsFlores.push({ tipo: nombreFinal, emoji: flor.emoji, colores: [...colores], cantidad });
    renderListaFlores();
    actualizarDescripcionCompilada();

    // Se reinician los pasos 1-3 para permitir agregar otro tipo de flor
    state.constructorFlor = { tipo: null, colores: [] };
    document.querySelectorAll('.tipo-flor-btn').forEach(btn => btn.classList.remove('border-rose-500', 'ring-2', 'ring-rose-200', 'bg-rose-50'));
    document.getElementById('panel-flor-personalizada').classList.add('hidden');
    document.getElementById('nombreFlorPersonalizada').value = '';
    document.getElementById('panel-color-flor').classList.add('hidden');
    document.getElementById('panel-cantidad-flor').classList.add('hidden');
    document.getElementById('cantidadFlorActual').value = 6;
    document.getElementById('colorPersonalizado').value = '';
}

function eliminarFlorDelPedido(index) {
    state.itemsFlores.splice(index, 1);
    renderListaFlores();
    actualizarDescripcionCompilada();
}

function renderListaFlores() {
    const cont = document.getElementById('lista-flores-pedido');
    const contador = document.getElementById('contadorFlores');
    const n = state.itemsFlores.length;
    contador.innerText = `${n} tipo${n === 1 ? '' : 's'} agregado${n === 1 ? '' : 's'}`;

    if (!n) {
        cont.innerHTML = '<p class="text-xs text-slate-400 italic text-center py-3">Aún no has agregado flores. Elige un tipo arriba para empezar 👆</p>';
        return;
    }

    cont.innerHTML = state.itemsFlores.map((item, i) => `
        <div class="flex items-center justify-between bg-white border border-rose-100 rounded-lg px-3 py-2 fade-in">
            <span class="text-sm text-slate-700"><span class="text-lg mr-1">${item.emoji}</span><strong>${item.cantidad}</strong> ${item.tipo} <span class="text-slate-300 mx-1">·</span> ${item.colores.join(', ')}</span>
            <button type="button" onclick="window.eliminarFlorDelPedido(${i})" class="text-rose-300 hover:text-red-500 transition-colors px-2"><i class="fas fa-times"></i></button>
        </div>
    `).join('');
}

function agregarDetalleRapido(texto) {
    const el = document.getElementById('detallesAdicionales');
    el.value = el.value.trim() ? `${el.value.trim()}. ${texto}` : texto;
    actualizarDescripcionCompilada();
}

function actualizarSaldoPendiente() {
    const total = parseFloat(document.getElementById('montoTotal').value) || 0;
    const anticipo = parseFloat(document.getElementById('montoAnticipo').value) || 0;
    const saldo = Math.max(total - anticipo, 0);
    document.getElementById('saldoPendiente').value = new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', minimumFractionDigits: 0 }).format(saldo);
}

function actualizarDescripcionCompilada() {
    const partesFlores = state.itemsFlores.map(i => `${i.cantidad} ${i.tipo} (${i.colores.join(', ')})`).join(', ');
    const detalles = document.getElementById('detallesAdicionales').value.trim();
    let texto = partesFlores;
    if (detalles) texto += (texto ? '. Detalles: ' : 'Detalles: ') + detalles;
    document.getElementById('descripcionPedido').value = texto;
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', inicializar);
} else {
    inicializar();
}

// --- FUNCIONES PRINCIPALES ---
async function cargarPedidos() {
    try {
        state.pedidos = await obtenerPedidos();
        state.pedidos.sort((a, b) => new Date(a.fechaEntrega) - new Date(b.fechaEntrega));
        aplicarFiltros();
        mostrarPanelNotificaciones();
    } catch (e) {
        console.error("Error cargando pedidos:", e);
        mostrarErrorCarga(e);
    }
}

function mostrarErrorCarga(e) {
    const lista = document.getElementById('lista-pedidos');
    if (!lista) return;
    const msg = (e && e.message) ? e.message : String(e);
    lista.innerHTML = `
        <div class="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            <p class="font-bold mb-1"><i class="fas fa-exclamation-triangle mr-1"></i> No se pudieron cargar los pedidos</p>
            <p class="text-xs text-red-600 font-mono break-all">${msg}</p>
            <p class="text-xs text-slate-500 mt-2">Verifica la conexión con el backend en Railway y la base de datos PostgreSQL.</p>
        </div>`;
}

async function enviarPedido() {
    const btn = document.querySelector('button[onclick="window.enviarPedido()"]');
    const textoOriginal = btn.innerText;
    btn.innerText = 'Enviando...';
    btn.disabled = true;

    // Validación
    const ids = ['nombreRemitente', 'celularRemitente', 'fechaEntrega', 'horaEntrega', 'pais', 'ciudad', 'barrio', 'nombreDestinatario', 'direccion', 'montoTotal', 'montoAnticipo'];
    let valido = true;
    
    ids.forEach(id => {
        const el = document.getElementById(id);
        if (!el.value.trim()) {
            el.style.borderColor = 'red';
            valido = false;
        } else {
            el.style.borderColor = '#e5e7eb';
        }
    });

    if (!valido) {
        alert('Faltan campos obligatorios');
        btn.innerText = textoOriginal;
        btn.disabled = false;
        return;
    }

    if (state.itemsFlores.length === 0) {
        alert('Agrega al menos un tipo de flor a tu pedido 💐');
        btn.innerText = textoOriginal;
        btn.disabled = false;
        document.getElementById('constructor-flores').scrollIntoView({ behavior: 'smooth', block: 'center' });
        return;
    }

    const data = {
        nombreRemitente: document.getElementById('nombreRemitente').value,
        celularRemitente: document.getElementById('celularRemitente').value,
        fechaEntrega: document.getElementById('fechaEntrega').value,
        horaEntrega: document.getElementById('horaEntrega').value,
        pais: document.getElementById('pais').value,
        ciudad: document.getElementById('ciudad').value,
        barrio: document.getElementById('barrio').value,
        nombreDestinatario: document.getElementById('nombreDestinatario').value,
        celularDestinatario: document.getElementById('celularDestinatario').value,
        direccion: document.getElementById('direccion').value,
        puntoReferencia: document.getElementById('puntoReferencia').value || '',
        descripcionPedido: document.getElementById('descripcionPedido').value,
        itemsFlores: state.itemsFlores,
        mensajeTarjeta: document.getElementById('mensajeTarjeta').value, // Aquí se captura
        montoTotal: document.getElementById('montoTotal').value,
        montoAnticipo: document.getElementById('montoAnticipo').value,
        metodoPagoAnticipo: document.getElementById('metodoPagoAnticipo').value,
        medioPago: document.querySelector('input[name="medioPago"]:checked').value,
        estado: 'pendiente',
        fechaCreacion: new Date().toISOString(),
        numeroWhatsAppUsado: CONFIG.NUMERO_WHATSAPP
    };

    try {
        await crearPedido(data);
        const fmtMoneda = (n) => new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', minimumFractionDigits: 0 }).format(n || 0);
        const monto = fmtMoneda(data.montoAnticipo);
        const total = fmtMoneda(data.montoTotal);
        const saldo = fmtMoneda(Math.max((Number(data.montoTotal) || 0) - (Number(data.montoAnticipo) || 0), 0));
        
        const msg = `🌸 *NUEVO PEDIDO KAREN'S FLORISTIK* 🌸%0A%0A*👤 De:* ${data.nombreRemitente}%0A*📅 Para:* ${data.fechaEntrega} a las ${data.horaEntrega}%0A%0A*📍 Destino:* ${data.nombreDestinatario}%0A*🗺️ Ubicación:* ${data.ciudad}, ${data.pais} - ${data.barrio}%0A*🏠 Dirección:* ${data.direccion}%0A%0A*💐 Pedido:* ${data.descripcionPedido}%0A💌 *Tarjeta:* ${data.mensajeTarjeta}%0A%0A*🧾 Total Pedido:* ${total}%0A*💰 Anticipo:* ${monto} (${data.metodoPagoAnticipo})%0A*⏳ Saldo Pendiente:* ${saldo}`;
        
        window.open(`https://wa.me/${CONFIG.NUMERO_WHATSAPP}?text=${msg}`, '_blank');
        
        window.mostrarExito();
        
        setTimeout(() => {
            document.getElementById('pedidoForm').reset();
            state.itemsFlores = [];
            state.constructorFlor = { tipo: null, colores: [] };
            renderListaFlores();
            document.getElementById('panel-color-flor').classList.add('hidden');
            document.getElementById('panel-cantidad-flor').classList.add('hidden');
            document.getElementById('chips-colores-seleccionados').innerHTML = '';
            document.getElementById('panel-flor-personalizada').classList.add('hidden');
            document.querySelectorAll('.tipo-flor-btn').forEach(b => b.classList.remove('border-rose-500', 'ring-2', 'ring-rose-200', 'bg-rose-50'));
            window.actualizarSaldoPendiente();
            window.mostrarFormulario();
            cargarPedidos();
        }, 5000);
    } catch (e) {
        console.error(e);
        alert(`Error al guardar el pedido: ${e && e.message ? e.message : 'intenta nuevamente.'}`);
    } finally {
        btn.innerText = textoOriginal;
        btn.disabled = false;
    }
}

function verificarPassword() {
    const pass = document.getElementById('password').value;
    if (pass === CONFIG.PASSWORD_ADMIN) {
        document.getElementById('login-view').classList.add('hidden');
        document.getElementById('admin-view').classList.remove('hidden');
        document.getElementById('password').value = '';
    } else {
        alert('Contraseña incorrecta');
    }
}

// --- ADMINISTRACIÓN Y RENDERIZADO DE TARJETAS ---
function crearTarjeta(p) {
    const fmt = (n) => new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', minimumFractionDigits: 0 }).format(n || 0);
    const monto = fmt(p.montoAnticipo);
    const hayTotal = p.montoTotal !== undefined && p.montoTotal !== null && p.montoTotal !== '';
    const montoTotalTexto = hayTotal ? fmt(p.montoTotal) : null;
    const saldoTexto = hayTotal ? fmt(Math.max((Number(p.montoTotal) || 0) - (Number(p.montoAnticipo) || 0), 0)) : null;
    
    // Configuración de colores
    let borderClass = 'border-l-4 border-gray-300';
    let bgClass = 'bg-white'; // Fondo base de la tarjeta
    let activePendiente = 'bg-white text-gray-600 border border-gray-200';
    let activeProceso = 'bg-white text-gray-600 border border-gray-200';
    let activeListo = 'bg-white text-gray-600 border border-gray-200';

    if(p.estado === 'pendiente') { 
        borderClass = 'border-l-4 border-yellow-400'; 
        activePendiente = 'bg-yellow-100 text-yellow-800 font-bold border-yellow-300';
    }
    else if(p.estado === 'proceso') { 
        borderClass = 'border-l-4 border-blue-400'; 
        activeProceso = 'bg-blue-100 text-blue-800 font-bold border-blue-300';
    }
    else if(p.estado === 'completado') { 
        borderClass = 'border-l-4 border-green-400'; 
        activeListo = 'bg-green-100 text-green-800 font-bold border-green-300';
    }

    // Datos procesados
    const telefono = p.celularDestinatario ? p.celularDestinatario : '<span class="text-gray-400 italic">Sin teléfono</span>';
    const ubicacionCompleta = [p.direccion, p.barrio, p.ciudad, p.pais].filter(Boolean).join(', ');
    const displayUbicacion = ubicacionCompleta || 'Sin dirección registrada';
    const mensajeTarjeta = p.mensajeTarjeta ? p.mensajeTarjeta : 'Sin mensaje para la tarjeta';
    const itemsFloresHtml = (Array.isArray(p.itemsFlores) && p.itemsFlores.length)
        ? `<div class="flex flex-wrap gap-1 mb-2">${p.itemsFlores.map(i => {
            const coloresTexto = Array.isArray(i.colores) ? i.colores.join(', ') : (i.color || '');
            return `<span class="bg-white border border-yellow-200 rounded-full px-2 py-0.5 text-xs text-yellow-700">${i.emoji || '🌸'} ${i.cantidad} ${i.tipo} · ${coloresTexto}</span>`;
        }).join('')}</div>`
        : '';

    return `
    <div class="rounded-xl shadow-sm hover:shadow-md transition-shadow p-5 mb-0 flex flex-col gap-3 relative bg-white ${borderClass}">
        
        <div class="flex justify-between items-start border-b border-gray-100 pb-2">
            <h3 class="font-bold text-lg text-gray-800 leading-tight">${p.nombreDestinatario}</h3>
            <button onclick="window.eliminarPedido('${p.id}')" class="text-gray-300 hover:text-red-500 transition-colors px-2">
                <i class="fas fa-trash-alt"></i>
            </button>
        </div>

        <div class="flex items-center text-sm text-gray-600 bg-gray-50 p-2 rounded-lg">
            <i class="fas fa-calendar-alt text-rose-400 mr-2 w-5 text-center"></i>
            <span class="font-medium">${p.fechaEntrega}</span>
            <span class="mx-2 text-gray-300">|</span>
            <span class="bg-rose-100 text-rose-700 px-2 py-0.5 rounded text-xs font-bold">${p.horaEntrega}</span>
        </div>

        <div class="bg-slate-50 p-3 rounded-lg border border-slate-100 text-sm">
            <div class="flex items-center mb-2 pb-2 border-b border-slate-100">
                <i class="fas fa-phone text-rose-400 mr-2 w-5 text-center"></i>
                <span class="font-semibold text-slate-700">${telefono}</span>
            </div>
            <div class="flex items-start">
                <i class="fas fa-map-marker-alt text-rose-400 mr-2 w-5 text-center mt-0.5"></i>
                <div>
                    <strong class="block text-xs text-rose-400 uppercase tracking-wide mb-0.5">Ubicación:</strong>
                    <p class="text-slate-600 leading-snug">${displayUbicacion}</p>
                </div>
            </div>
        </div>

        <div class="text-sm bg-yellow-50 p-3 rounded border border-yellow-100">
             <strong class="block text-xs text-yellow-600 uppercase mb-1">💐 Pedido:</strong>
             ${itemsFloresHtml}
             <span class="text-gray-700 italic">"${p.descripcionPedido}"</span>
        </div>

        <div class="text-sm bg-pink-50 p-3 rounded border border-pink-100">
             <strong class="block text-xs text-pink-500 uppercase mb-1">💌 Mensaje Tarjeta:</strong>
             <span class="text-gray-700 italic font-serif">"${mensajeTarjeta}"</span>
        </div>

        <div class="text-sm bg-green-50 p-3 rounded-lg border border-green-100 space-y-1.5">
            <div class="flex items-center gap-2 text-gray-700">
                <i class="fas fa-money-bill-wave text-green-500 w-5 text-center"></i>
                ${montoTotalTexto ? `<span>Total: <span class="font-bold text-gray-800">${montoTotalTexto}</span></span>` : '<span class="text-xs text-gray-400 italic">Total no registrado (pedido anterior)</span>'}
            </div>
            <div class="flex items-center justify-between pl-7">
                <span class="text-gray-700">Anticipo: <span class="font-bold text-green-700">${monto}</span> <span class="text-xs text-gray-500">(${p.metodoPagoAnticipo})</span></span>
            </div>
            ${saldoTexto ? `<div class="flex items-center justify-between pl-7"><span class="text-gray-700">Saldo: <span class="font-bold text-rose-600">${saldoTexto}</span></span></div>` : ''}
        </div>

        <div class="flex rounded-md overflow-hidden border border-gray-200 mt-2 shadow-sm">
            <button onclick="window.cambiarEstado('${p.id}', 'pendiente')" class="flex-1 text-xs py-2 transition-colors ${activePendiente}">Pendiente</button>
            <button onclick="window.cambiarEstado('${p.id}', 'proceso')" class="flex-1 text-xs py-2 transition-colors ${activeProceso}">Proceso</button>
            <button onclick="window.cambiarEstado('${p.id}', 'completado')" class="flex-1 text-xs py-2 transition-colors ${activeListo}">Listo</button>
        </div>
    </div>`;
}

function aplicarFiltros() {
    const term = document.getElementById('searchTerm').value.toLowerCase();
    const fDate = document.getElementById('filtroFecha').value;
    const fState = document.getElementById('filtroEstado').value;
    const today = new Date().toISOString().split('T')[0];

    const filtrados = state.pedidos.filter(p => {
        const fullText = JSON.stringify(p).toLowerCase();
        const matchText = fullText.includes(term);
        const matchState = fState === 'todos' || (fState === 'pendientes' ? (!p.estado || p.estado === 'pendiente') : p.estado === fState);
        let matchDate = true;
        if(fDate === 'hoy') matchDate = p.fechaEntrega === today;
        if(fDate === 'proximos') matchDate = p.fechaEntrega > today;
        return matchText && matchState && matchDate;
    });

    const lista = document.getElementById('lista-pedidos');
    if (!filtrados.length) {
        lista.innerHTML = '<div class="text-center py-10 text-gray-400 bg-white rounded-xl border border-dashed border-gray-200">No se encontraron pedidos con estos filtros</div>';
        return;
    }

    // Ya vienen ordenados ascendente por fechaEntrega desde cargarPedidos(); agrupamos por mes.
    const haceUnMes = new Date();
    haceUnMes.setDate(haceUnMes.getDate() - 30);
    const inicioMesLimite = new Date(haceUnMes.getFullYear(), haceUnMes.getMonth(), 1);

    const grupos = {};
    filtrados.forEach(p => {
        const fecha = p.fechaEntrega ? new Date(p.fechaEntrega + 'T00:00:00') : null;
        const key = (fecha && !isNaN(fecha)) ? `${fecha.getFullYear()}-${String(fecha.getMonth() + 1).padStart(2, '0')}` : 'sin-fecha';
        (grupos[key] = grupos[key] || []).push(p);
    });

    lista.innerHTML = Object.keys(grupos).sort().map(key => {
        const pedidosGrupo = grupos[key];
        const fechaGrupo = key !== 'sin-fecha' ? new Date(key + '-01T00:00:00') : null;
        const esAntiguo = fechaGrupo ? fechaGrupo < inicioMesLimite : false;

        if (!state.gruposInicializados.has(key)) {
            state.gruposInicializados.add(key);
            if (esAntiguo) state.gruposColapsados.add(key);
        }
        const colapsado = state.gruposColapsados.has(key);
        const nombreMes = fechaGrupo
            ? fechaGrupo.toLocaleDateString('es-CO', { month: 'long', year: 'numeric' })
            : 'Sin fecha de entrega';

        return `
        <div>
            <button type="button" onclick="window.toggleGrupoFecha('${key}')" class="w-full flex items-center justify-between bg-rose-100/70 hover:bg-rose-100 rounded-xl px-4 py-2.5 mb-3 transition-colors">
                <span class="font-bold text-rose-800 capitalize flex items-center gap-2">
                    <i class="fas fa-chevron-${colapsado ? 'right' : 'down'} text-xs transition-transform"></i>
                    ${nombreMes}
                    ${esAntiguo ? '<span class="text-[10px] font-normal normal-case text-rose-400">(hace más de un mes)</span>' : ''}
                </span>
                <span class="text-xs font-bold text-rose-500 bg-white px-2 py-1 rounded-full">${pedidosGrupo.length} pedido${pedidosGrupo.length === 1 ? '' : 's'}</span>
            </button>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-6 ${colapsado ? 'hidden' : ''}">
                ${pedidosGrupo.map(crearTarjeta).join('')}
            </div>
        </div>`;
    }).join('');
}

function toggleGrupoFecha(key) {
    if (state.gruposColapsados.has(key)) state.gruposColapsados.delete(key);
    else state.gruposColapsados.add(key);
    aplicarFiltros();
}

function mostrarPanelNotificaciones() {
    const hoy = new Date().toISOString().split('T')[0];
    const urgentes = state.pedidos.filter(p => p.fechaEntrega === hoy && p.estado !== 'completado');
    const panel = document.getElementById('panel-notificaciones-admin');
    if(urgentes.length > 0) {
        panel.innerHTML = `
            <div class="bg-rose-500 text-white p-4 rounded-xl shadow-lg animate-pulse flex items-center justify-center gap-3 mb-4">
                <i class="fas fa-bell text-xl"></i>
                <span class="font-bold text-lg">¡ATENCIÓN! Tienes ${urgentes.length} entregas pendientes para HOY</span>
            </div>`;
    } else {
        panel.innerHTML = '';
    }
}

async function cambiarEstado(id, nuevoEstado) {
    await actualizarEstadoPedido(id, nuevoEstado);
    cargarPedidos();
}

function eliminarPedido(id) {
    state.pedidoAEliminar = id;
    document.getElementById('modal-eliminar').classList.remove('hidden');
}

async function confirmarEliminacion() {
    if(state.pedidoAEliminar) {
        await apiEliminarPedido(state.pedidoAEliminar);
        document.getElementById('modal-eliminar').classList.add('hidden');
        cargarPedidos();
    }
}

function exportarExcel() {
    const encabezados = [
        'Estado', 'Fecha Entrega', 'Hora Entrega', 'Destinatario', 'Tel. Destinatario',
        'Dirección', 'Barrio', 'Ciudad', 'País', 'Punto Referencia',
        'Flores del Pedido', 'Descripción Completa', 'Mensaje Tarjeta',
        'Total (COP)', 'Anticipo (COP)', 'Saldo Pendiente (COP)', 'Método Anticipo',
        'Remitente', 'Tel. Remitente', 'Fecha Creación Pedido'
    ];

    // Escapa una celda para CSV: si contiene coma, comillas o salto de línea, la envuelve en comillas
    const escaparCelda = (valor) => {
        const texto = (valor === undefined || valor === null) ? '' : String(valor);
        return /[",\n]/.test(texto) ? `"${texto.replace(/"/g, '""')}"` : texto;
    };

    const pedidosOrdenados = [...state.pedidos].sort((a, b) => new Date(a.fechaEntrega) - new Date(b.fechaEntrega));

    const filas = pedidosOrdenados.map(p => {
        const total = Number(p.montoTotal) || 0;
        const anticipo = Number(p.montoAnticipo) || 0;
        const saldo = Math.max(total - anticipo, 0);
        const floresTexto = (Array.isArray(p.itemsFlores) && p.itemsFlores.length)
            ? p.itemsFlores.map(i => {
                const colores = Array.isArray(i.colores) ? i.colores.join('/') : (i.color || '');
                return `${i.cantidad} ${i.tipo} (${colores})`;
            }).join(' + ')
            : '';

        return [
            (p.estado || 'pendiente').toUpperCase(),
            p.fechaEntrega || '',
            p.horaEntrega || '',
            p.nombreDestinatario || '',
            p.celularDestinatario || '',
            p.direccion || '',
            p.barrio || '',
            p.ciudad || '',
            p.pais || '',
            p.puntoReferencia || '',
            floresTexto,
            p.descripcionPedido || '',
            p.mensajeTarjeta || '',
            total,
            anticipo,
            saldo,
            p.metodoPagoAnticipo || '',
            p.nombreRemitente || '',
            p.celularRemitente || '',
            p.fechaCreacion ? new Date(p.fechaCreacion).toLocaleString('es-CO') : ''
        ];
    });

    const lineas = [encabezados, ...filas].map(fila => fila.map(escaparCelda).join(','));
    // \uFEFF (BOM) para que Excel detecte UTF-8 y no rompa tildes/ñ
    const csv = '\uFEFF' + lineas.join('\r\n');

    const a = document.createElement('a');
    a.href = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8;' }));
    a.download = `Pedidos_KarensFloristik_${new Date().toISOString().slice(0,10)}.csv`;
    a.click();
}

// --- EXPOSICIÓN GLOBAL ---
window.verificarPassword = verificarPassword;
window.enviarPedido = enviarPedido;
window.aplicarFiltros = aplicarFiltros;
window.toggleGrupoFecha = toggleGrupoFecha;
window.cambiarEstado = cambiarEstado;
window.eliminarPedido = eliminarPedido;
window.confirmarEliminacion = confirmarEliminacion;
window.cancelarEliminacion = () => document.getElementById('modal-eliminar').classList.add('hidden');
window.exportarExcel = exportarExcel;
window.seleccionarTipoFlor = seleccionarTipoFlor;
window.seleccionarColorFlor = seleccionarColorFlor;
window.cambiarCantidadFlor = cambiarCantidadFlor;
window.agregarFlorAlPedido = agregarFlorAlPedido;
window.eliminarFlorDelPedido = eliminarFlorDelPedido;
window.agregarDetalleRapido = agregarDetalleRapido;
window.actualizarDescripcionCompilada = actualizarDescripcionCompilada;
window.agregarColorPersonalizado = agregarColorPersonalizado;
window.eliminarColorPersonalizado = eliminarColorPersonalizado;
window.actualizarSaldoPendiente = actualizarSaldoPendiente;