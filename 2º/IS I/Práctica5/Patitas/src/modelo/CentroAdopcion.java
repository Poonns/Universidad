
package modelo;
import java.util.ArrayList;


public class CentroAdopcion {
	private ArrayList<Personal> personal = new ArrayList<Personal>();
	private ArrayList<Adoptante> adoptantes = new ArrayList<Adoptante>();
	private ArrayList<SolicitudAdopcion> solicitudesAdopcion = new ArrayList<SolicitudAdopcion>();
	private ArrayList<Animal> animales = new ArrayList<Animal>();
	private ArrayList<Visita> visitas = new ArrayList<Visita>();

	public CentroAdopcion() {
		
	}
        
        /**
         * 
         * @author Arian Gisbert, Hector Pons
         * 
         * @return 
         */
        public ArrayList<Animal> getAnimalesDisponibles(){
            
           
           ArrayList<Animal> listaAnimalesDisponibles = new ArrayList<Animal>(animales);
            
           for(Animal animal : animales){
               
               //Si el animal no esta disponible
               if(!animal.comprobarDisponibilidadAdopcion()){
                   
                   listaAnimalesDisponibles.remove(animal);
                   
                   
               }
               
           } 
           
           return listaAnimalesDisponibles;
           
            
        }

    public void setAdoptantes(ArrayList<Adoptante> adoptantes) {
        this.adoptantes = adoptantes;
    }

    public void setAnimales(ArrayList<Animal> animales) {
        this.animales = animales;
    }
    
    public void addSolicitud(SolicitudAdopcion solicitud){
        
        solicitudesAdopcion.add(solicitud);
        
        System.out.println("Solicitud creada correctamente."+solicitud.toString());
        
    }
        
        
        
}