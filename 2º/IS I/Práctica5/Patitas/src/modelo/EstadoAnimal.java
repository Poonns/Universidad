package modelo;
public abstract class EstadoAnimal {
    
protected Animal animal;


        public String toDisponible(){
		animal.cambioEstado(new NoAdoptable(animal));
		return "NoAdoptable";
	}
	

        public String toNoAdoptable(){
		animal.cambioEstado(new NoAdoptable(animal));
		return "NoAdoptable";
	}
	
	public String toBajoTratamiento(){
		animal.cambioEstado(new BajoTratamiento(animal));
		return "Bajo Tratamiento";
	}
	
	public String toProcesoAdopcion(){
		animal.cambioEstado(new ProcesoAdopcion(animal));
		return "ProcesoAdopcion";
	}
	
	public String toAdoptado(){
		 animal.cambioEstado(new Adoptado(animal));
		return "Adoptado";
	}

}