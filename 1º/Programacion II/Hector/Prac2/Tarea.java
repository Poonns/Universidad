package Hector.Prac2;

public class Tarea {
    private Fecha fecha;
    private String descripcion;

    public Tarea(Fecha fecha, String descripcion){
        this.fecha = fecha;
        this.descripcion = descripcion;
    }

    public Fecha getfecha(){
        return this.fecha;
    }

    public String getDescripcion(){
        return this.descripcion;
    }

    public String toString(){
        return this.fecha + ": " + this.descripcion;
    }
    
}

public class Agenda{
    private Tarea[] tareas;

    public Agenda(){
        tareas = new Tarea[0];
    }

    private int posicionInsercion(Tarea tarea){
        int pos = 0;
        for(int i = 0; i < tareas.length; i++){
            if(tarea.getfecha().compareTo(tareas[i].getfecha()) <= 0){
                pos = i;
            }
        }
        return pos;
    }

    public void añadir(Tarea tarea){
        int pos = posicionInsercion(tarea);

        Tarea[] nuevoVector = new Tarea[tareas.length + 1];

        for(int i = 0; i < pos; i++){
            nuevoVector[i] = tareas[i];
        }

        for(int i = pos; i < nuevoVector.length; i++){
            nuevoVector[i] = tareas[i];
        }

        nuevoVector = tareas;
    }

    public int cantidad(){
        return tareas.length;
    }

    public Tarea[] consultar(Fecha fecha){
        int contador = 0;
        for(int i = 0; i < tareas.length; i++){
            if(tareas[i].getfecha().equals(fecha)){
                contador ++;
            }
        }

        if(contador == 0){
            Tarea[] nuevoVector = new Tarea[0];
        }

        Tarea[] nuevoVector = new Tarea[contador];
        int j = 0;

        for(int i = 0; i < tareas.length; i++){
            nuevoVector[j] = tareas[i];
            j++;
        }

        return nuevoVector;
    }

    public void borrarPasadas(Fecha fecha){
        int eliminar = 0;
        for(int i = 0; i < tareas.length; i++){
            if(fecha.compareTo(tareas[i].getfecha()) > 0){
                eliminar++;
            }
        }

        Tarea[] nuevoVector = new Tarea[tareas.length - eliminar];
        int j = 0;

        for(int i = 0; i < tareas.length; i++){
            if(fecha.compareTo(tareas[i].getfecha()) <= 0){
                nuevoVector[j] = tareas[i];
                j++;
            }
        }
    }

    public void borrar(){
        Fecha hoy = new Fecha(10, 2, 2002);
        int eliminar = 0;

        for(int i = 0; i < tareas.length; i++){
            if(hoy.compareTo(tareas[i].getfecha()) > 0){
                eliminar++;
            }
        }
        
        Tarea[] nuevoVector = new Tarea[tareas.length - eliminar];
        int j = 0;

        for(int i = 0; i < tareas.length; i++){
            if(hoy.compareTo(tareas[i].getfecha()) <= 0){
                nuevoVector[j] = tareas[i];
                j++;
            }
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