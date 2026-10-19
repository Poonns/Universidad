package modelo;
import java.util.Vector;
import java.util.Date;
import java.util.ArrayList;

public class Animal {
	private String identificador;
	private String especie;
	private String raza;
	private int edad;
	private String sexo;
	private Date fechaIngreso;
	private String observaciones;
	private CentroAdopcion centroAdopcion;
	private ArrayList<Visita> visitas = new ArrayList<Visita>();
	private ArrayList<SolicitudAdopcion> solicitudAdopcion = new ArrayList<SolicitudAdopcion>();
	private EstadoAnimal estadoAnimal;
	private ArrayList<Tratamiento> tratamientos = new ArrayList<Tratamiento>();

	public Animal() {
		throw new UnsupportedOperationException();
	}
        
        //Constructor para la creación inicial
        
        public Animal(String identificador, String especie, String raza, int edad, Date fechaIngreso, String observaciones, CentroAdopcion centroAdopcion, EstadoAnimal estadoAnimal){
            
            this.identificador = identificador;
            this.especie = especie;
            this.raza = raza;
            this.edad = edad;
            this.fechaIngreso = fechaIngreso;
            this.observaciones = observaciones;
            this.centroAdopcion = centroAdopcion;
            this.estadoAnimal = estadoAnimal;
               
            
        }
        
        public void cambioEstado(EstadoAnimal nuevoEstado){
            
            this.estadoAnimal = nuevoEstado;
            
       
        }
        
        /**
         * 
         * @author Arian Gisbert, Hector Pons
         * 
         * @return 
         */
        public boolean comprobarDisponibilidadAdopcion(){
            
            if(estadoAnimal instanceof DisponibleAdopcion){
                
                return true;
                
            }
            
            return false;
            
        }
        
        @Override
        public String toString(){
            
           return "Especie: "+this.especie +" | Raza: "+this.raza+" | Edad: "+this.edad;
            
        }
        
        
        
}