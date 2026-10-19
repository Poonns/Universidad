package modelo;
public class Cancelada extends EstadoVisita {
    private Visita v;
 
    public Cancelada(Visita visita) {
		this.v = visita;
		v.setEstadoVisita(this);
	}

	public Cancelada crear(Visita vr) {
		this.v = vr;
		if (v.getEstadoVisita().getNombre().equals("Pendiente")) {
			v.setEstadoVisita(new Cancelada(vr));
			// Aquí podrías registrar el motivo de rechazo si lo deseas
			return this;
		} else {
			throw new IllegalStateException("Solo se pueden cancelar visitas pendientes.");
		}
	}
 
    @Override
    public String confirmar(Visita visita) {
		return "No se puede confirmar una visita ya cancelada.";
    }
 
    @Override
    public String rechazar(Visita visita, String motivo) {
        return "La visita ya está rechazada.";
    }
	 @Override
    public String cancelar(Visita visita) {
        return "La visita ya está Cancelada.";
    }
    @Override
    public String getNombre() {
        return "Cancelada";
    }
}