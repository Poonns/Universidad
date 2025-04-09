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

            Tarea[] nuevoVector = new Tarea[tareas.length + 1];

            for(int i = 0; i < posicion; i++){
                nuevoVector[i] = tareas[i];
            }

            nuevoVector[posicion] = tarea;

            for(int i = posicion; i < tareas.length; i++){
                nuevoVector[i+1] = tareas[i];
            }

            nuevoVector = tareas;
        }

        public int cantidad(){
            return tareas.length;
        }

        public Tarea[] consultar(Fecha fecha){
            int contador = 0;

            for(int i = 0; i < tareas.length; i++){
                if(fecha.equals(tareas[i].fecha)){
                    contador ++;
                }
            }

            if(contador == 0){
                return new Tarea[0];
            }

            Tarea[] nuevoVector = new Tarea[contador];
            int i = 0;

            for(int j = 0; j < nuevoVector.length; j++){
                if(fecha.equals(tareas[j].fecha)){
                    nuevoVector[i] = tareas[j];
                    i++;
                }
            }

            return nuevoVector;
        }

        public void borrarPasadas(Fecha fecha){
            int eliminar = 0;
            
            for(int i = 0; i < tareas.length; i++){
                if(fecha.año < tareas[i].fecha.año || fecha.año == tareas[i].fecha.año && fecha.mes < tareas[i].fecha.mes || fecha.año == tareas[i].fecha.año && fecha.mes == tareas[i].fecha.mes && fecha.dia < tareas[i].fecha.dia){
                    eliminar++;
                }
            }
            int dimensionVector = tareas.length - eliminar;
            Tarea[] nuevoVector = new Tarea[dimensionVector];
            int j = 0;

            for(int i = 0; i < tareas.length; i++){
                if(fecha.año >= tareas[i].fecha.año || fecha.año == tareas[i].fecha.año && fecha.mes >= tareas[i].fecha.mes || fecha.año == tareas[i].fecha.año && fecha.mes == tareas[i].fecha.mes && fecha.dia >= tareas[i].fecha.dia){
                    nuevoVector[j] = tareas[i];
                    j++;
                }
            }

            tareas = nuevoVector;
        }

        public void borrar(){
            Fecha hoy = new Fecha(9, 4, 2025);

            borrarPasadas(hoy);
        }

        public String toString(){
            String resultado = "";

            for(int i = 0; i < tareas.length; i++){
                resultado += tareas[i].toString() + "\n";
            }
            return resultado;
        }
    }
}

