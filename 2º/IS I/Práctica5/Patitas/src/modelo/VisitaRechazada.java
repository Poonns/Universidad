package modelo;
/**
 * Estado concreto: la visita ha sido rechazada.
 * Estado terminal: no admite más transiciones.
 */
public class VisitaRechazada extends EstadoVisita {

    private static final VisitaRechazada INSTANCIA = new VisitaRechazada();

    private VisitaRechazada() {}

    public static VisitaRechazada getInstancia() {
        return INSTANCIA;
    }

    @Override
    public void confirmar(Visita visita) {
        throw new IllegalStateException("No se puede confirmar una visita ya rechazada.");
    }

    @Override
    public void rechazar(Visita visita, String motivo) {
        throw new IllegalStateException("La visita ya está rechazada.");
    }

    @Override
    public String getNombre() {
        return "Rechazada";
    }
}