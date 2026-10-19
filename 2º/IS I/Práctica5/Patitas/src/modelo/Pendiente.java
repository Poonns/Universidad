package modelo;
/**
 * Estado concreto: la visita está pendiente de revisión.
 * Desde aquí se puede confirmar o rechazar.
 */
public class Pendiente extends EstadoVisita {

    private static final Pendiente INSTANCIA = new Pendiente();

    private Pendiente() {}

    public static Pendiente getInstancia() {
        return INSTANCIA;
    }

    @Override
    public void confirmar(Visita visita) {
        visita.setEstadoVisita(Confirmada.getInstancia());
    }

    @Override
    public void rechazar(Visita visita, String motivo) {
        if (motivo == null || motivo.trim().isEmpty()) {
            throw new IllegalArgumentException("El motivo de rechazo no puede estar vacío.");
        }
        visita.setMotivoRechazo(motivo);
        visita.setEstadoVisita(VisitaRechazada.getInstancia());
    }

    @Override
    public String getNombre() {
        return "Pendiente";
    }
}