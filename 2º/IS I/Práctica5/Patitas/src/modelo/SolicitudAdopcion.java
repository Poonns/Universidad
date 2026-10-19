package modelo;
import java.util.Date;

public class SolicitudAdopcion {
	private Date fecha;
	private int numPersonasDomicilio;
	private boolean ninyosDomicilio;
	private CentroAdopcion centroAdopcion;
	private Animal animal;
	private Adoptante adoptante;
	private EstadoAdopcion estadoAdopcion;
        private String disponibilidadHoraria;

	public SolicitudAdopcion(Animal animal,Adoptante adoptante, CentroAdopcion centroAdopcion, int numeroPersonas, boolean ninyosDomicilio, String disponibilidadHoraria ) {
		
            this.animal = animal;
            this.adoptante = adoptante;
            this.centroAdopcion = centroAdopcion;
            this.numPersonasDomicilio = numeroPersonas;
            this.ninyosDomicilio = ninyosDomicilio;
            this.disponibilidadHoraria = disponibilidadHoraria;
            
            
            //La fecha li posem la data de ara mateixa
            this.fecha = new Date(System.currentTimeMillis());
            
            //I marquem el estat de adopcion com en curso
            this.estadoAdopcion = new EnCurso();
            
	}

    @Override
    public String toString() {
        return "SolicitudAdopcion{" + "fecha=" + fecha + ", numPersonasDomicilio=" + numPersonasDomicilio + ", ninyosDomicilio=" + ninyosDomicilio + ", centroAdopcion=" + centroAdopcion + ", animal=" + animal + ", adoptante=" + adoptante + ", estadoAdopcion=" + estadoAdopcion + ", disponibilidadHoraria=" + disponibilidadHoraria + '}';
    }
        
        
        
        
}