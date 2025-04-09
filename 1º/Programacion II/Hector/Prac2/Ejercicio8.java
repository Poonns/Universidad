package Hector.Prac2;

public class Ejercicio8 {
    public class Fecha{

        private int dia;
        private int mes;
        private int año;

        public Fecha(int dia, int mes, int año){

            this.dia = dia;
            this.mes = mes;
            this.año = año;

        }
    
    }
    public class Tarea{

        private Fecha fecha;
        private String descripcion;

        public Tarea(Fecha fecha, String descripcion){
            
            this.fecha = fecha;
            this.descripcion = descripcion;

        }

        public Fecha getFecha(){
            return this.fecha;
        }

        public String getDescripcion(){
            return this.descripcion;
        }

        public String toString(){
            return fecha.toString() + ": " + descripcion;
        }
    }

    public class Agenda{
        private Tarea[] tareas;

        public Agenda(){
            tareas = new Tarea[0];
        }

        private int posicionInsercion(Tarea tarea) {
            Fecha fechaNueva = tarea.getFecha();
        
            for (int i = 0; i < tareas.length; i++) {
                Fecha actual = tareas[i].getFecha();

                if (fechaNueva.año < actual.año) {
                    return i; 
                }
        
                if (fechaNueva.año == actual.año) {
                    if (fechaNueva.mes < actual.mes) {
                        return i; 
                    }
        
                    if (fechaNueva.mes == actual.mes) {
                        if (fechaNueva.dia < actual.dia) {
                            return i;
                        }
                    }
                }
            }

            return tareas.length;
        }

        public void añadir(Tarea tarea){
            int posicion = posicionInsercion(tarea);

            Tarea[] nuevoVector = new Tarea[tarea.length + 1];

            for(int i = 0; i < tareas.length; i++){

            }
        }


    }
}

