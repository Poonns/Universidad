package modelo;
/**
 * Clase abstracta del patrón Estado para Visita.
 * Transiciones válidas:
 *   Pendiente → Confirmada  (confirmar)
 *   Pendiente → Rechazada   (rechazar)
 * Pendiente → Cancelada   (rechazar)
 * Confirmada, Cancelada y Rechazada son estados terminales.
 */
public abstract class EstadoVisita {
 
    public abstract void confirmar(Visita visita);
	public abstract void cancelar(Visita visita);
    public abstract void rechazar(Visita visita, String motivo);
 
    public abstract String getNombre();
}
 