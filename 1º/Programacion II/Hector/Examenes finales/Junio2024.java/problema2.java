package Hector.Examenes finales.Junio2024.java;

public class problema2 {
    public class EstudiantesUniversidad {
        public class Estudiante {
            private String dni;
            private boolean graduado;
    
            // Constructor
            public Estudiante(String dni, boolean graduado) {
                this.dni = dni;
                this.graduado = graduado;
            }
    
            // Getter del DNI
            public String getDni() {
                return dni;
            }
    
            // Getter de estado de graduación
            public boolean getGraduado() {
                return graduado;
            }
        }
        private class Nodo {
            Estudiante estudiante;
            Nodo sig;
            Nodo(Estudiante estudiante, Nodo sig) {
                this.estudiante = estudiante;
                this.sig = sig;
            }
        }
        // Atributos
        private String universidad; // Nombre de la universidad
        private Nodo primero; // Enlace al primer nodo de la lista de estudiantes
        // (nodos ordenados de menor a mayor por DNI)
    // Constructor
        public EstudiantesUniversidad(String universidad) {
            this.universidad = universidad;
            this.primero = null;
        }
        public boolean matricular(Estudiante estudiante){
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.estudiante.getDni().equals(estudiante.getDni())){
                    return false;
                } 
            }

            Nodo nuevo = new Nodo(estudiante, null);
            
            if(primero == null || estudiante.getDni().compareTo(primero.estudiante.getDni()) < 0){
                nuevo.sig = primero;
                primero = nuevo;
                return true;
            }
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                 if(actual.sig == null || actual.sig.estudiante.getDni().compareTo(estudiante.getDni()) > 0){
                    nuevo.sig = actual.sig;
                    actual.sig = nuevo;
                    return true;
                 } 
            }
            return false;
        }

        int eliminarGraduados(){
            int contador = 0;
            Nodo anterior = null;
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.estudiante.getGraduado()){
                    contador++;
                    if(anterior == null){
                        primero = actual.sig;
                    } else {
                        anterior.sig = actual.sig;
                    }
                } else{
                    anterior = actual;
                }
            }
            return contador;
        }
    }
}
