package modelo;
import java.util.Date;

public class Visita {
	/**
	 * @param idVisita El identificador de la Visita
	 * @param fecha La fecha de la visita. Estará en Formato Datetime	
	 * @param  Animal El Animal a visitar. Claramente,tomado por el ID
	 * @param cuidador El encargado de acompañar a la persona a la Visita.
	 * @param EstadoVisita El estado de la visita
	 * @param Adoptante La persona que le quiere adoptar
	 */
	private int idVisita;
	private Date fecha;
	private int animal;
	private int cuidador;o qu 
	private EstadoVisita estadoVisita;
	private int adoptante;
	private EstadoVisita EstadoVisita;
	private CentroAdopcion CentroAdopcion;

	/**
	 * Devuelve una lista de las Visita que estén pendientes
	 * @return ArrayList<Visita>
	 */
	protected void solicitarVisitasPendientes() {
		
	}

	public int getIdVisita() {
		return this.idVisita;
	}

	public void setFecha(Date f) {
		this.fecha = f;
	}

	public Date getFecha() {
		return this.fecha;
	}
	public String getFechaS() {
		return this.fecha.toString();
	}

	public void setAnimal(int a) {
		this.animal = a;
	}

	public int getAnimal() {
		return this.animal;
	}

	public void setCuidador(int _cuidador) {
		this.cuidador = _cuidador;
	}

	public int getCuidador() {
		return this.cuidador;
	}

	public void setEstadoVisita(EstadoVisita estadoVisita) {
		this.estadoVisita = estadoVisita;
	}

	public EstadoVisita getEstadoVisita() {
		return this.estadoVisita;
	}

	public int getAdoptante() {
		return this.adoptante;
	}
	    // ── toString ─────────────────────────────────────────────────────────────
 
    @Override
    public String toString() {
        return "Visita{id=" + idVisita
                + ", fecha=" + fecha
                + ", animal=" + animal
                + ", cuidador=" + cuidador
                + ", estado=" + estadoVisita.getNombre()
                + (motivoRechazo != null ? ", motivo='" + motivoRechazo + "'" : "")
                + "}";
    }
}