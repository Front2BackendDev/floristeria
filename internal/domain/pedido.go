package domain

// ItemFlor representa una flor configurada dentro del pedido.
type ItemFlor struct {
	Tipo     string   `json:"tipo"`
	Emoji    string   `json:"emoji,omitempty"`
	Colores  []string `json:"colores,omitempty"`
	Cantidad int      `json:"cantidad"`
}

// Pedido representa la orden completa registrada por un cliente.
type Pedido struct {
	ID                  string     `json:"id"`
	NombreRemitente     string     `json:"nombreRemitente"`
	CelularRemitente    string     `json:"celularRemitente"`
	FechaEntrega        string     `json:"fechaEntrega"`
	HoraEntrega         string     `json:"horaEntrega"`
	Pais                string     `json:"pais"`
	Ciudad              string     `json:"ciudad"`
	Barrio              string     `json:"barrio"`
	NombreDestinatario  string     `json:"nombreDestinatario"`
	CelularDestinatario string     `json:"celularDestinatario"`
	Direccion           string     `json:"direccion"`
	PuntoReferencia     string     `json:"puntoReferencia"`
	DescripcionPedido   string     `json:"descripcionPedido"`
	ItemsFlores         []ItemFlor `json:"itemsFlores,omitempty"`
	MensajeTarjeta      string     `json:"mensajeTarjeta"`
	MontoTotal          string     `json:"montoTotal"`
	MontoAnticipo       string     `json:"montoAnticipo"`
	MetodoPagoAnticipo  string     `json:"metodoPagoAnticipo"`
	MedioPago           string     `json:"medioPago"`
	Estado              string     `json:"estado"`
	FechaCreacion       string     `json:"fechaCreacion"`
	NumeroWhatsAppUsado string     `json:"numeroWhatsAppUsado"`
}

const (
	EstadoPendiente  = "pendiente"
	EstadoEnProceso  = "en_proceso"
	EstadoCompletado = "completado"
	EstadoCancelado  = "cancelado"
)
