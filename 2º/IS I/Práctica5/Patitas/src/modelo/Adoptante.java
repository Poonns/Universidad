package modelo;
import java.util.ArrayList;


public class Adoptante {
	private String nombre;
	private String dni;
	private String direccion;
	private String tipoVivienda;
	private String experienciaAnimales;
	private CentroAdopcion centroAdopcion;
	private ArrayList<SolicitudAdopcion> solicitudesAdopcion = new ArrayList<SolicitudAdopcion>();
	private ArrayList<Visita> visitas = new ArrayList<Visita>();

	public Adoptante() {
		throw new UnsupportedOperationException();
	}

        public Adoptante(String nombre, String dni, String direccion, String tipoVivienda, String experienciaAnimales, CentroAdopcion centroAdopcion) {
            this.nombre = nombre;
            this.dni = dni;
            this.direccion = direccion;
            this.tipoVivienda = tipoVivienda;
            this.experienciaAnimales = experienciaAnimales;
            this.centroAdopcion = centroAdopcion;
        }
        
        
        
        
        /**
         * 
         * @autor Arian Gisbert, Hector Pons
         *
         * Funció definida en el DSGS
         * 
         * @return 
         */
        public ArrayList<Animal> consultarAnimalesDisponibles(){
            
            return centroAdopcion.getAnimalesDisponibles();
            
        }
        
          /**
         * 
         * @autor Arian Gisbert, Hector Pons
         *
         * Funció definida en el DSGS
         * 
         * @return 
         */
        
        public void solicitarAdopcion(Object animal, int numeroPersonas, boolean presenciaNinos, String disponibilidadHoraria){
            
            //Creem la solicitud
            SolicitudAdopcion solicitud = new SolicitudAdopcion((Animal) animal, this, centroAdopcion, numeroPersonas, presenciaNinos, disponibilidadHoraria);
            
            centroAdopcion.addSolicitud(solicitud);
            this.solicitudesAdopcion.add(solicitud);
            
            
            
        }
        
        
}