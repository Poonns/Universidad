package modelo;

/**
 * Estado concreto: la visita ha sido confirmada.
 * Estado terminal: no admite más transiciones.
 */
public class Confirmada extends EstadoVisita {

    private static final Confirmada INSTANCIA = new Confirmada();

    private Confirmada() {}

    public static Confirmada getInstancia() {
        return INSTANCIA;
    }

    @Override
    public void confirmar(Visita visita) {
        throw new IllegalStateException("La visita ya está confirmada.");
    }

    @Override
    public void rechazar(Visita visita, String motivo) {
        throw new IllegalStateException("No se puede rechazar una visita ya confirmada.");
    }

    @Override
    public String getNombre() {
        return "Confirmada";
    }
}