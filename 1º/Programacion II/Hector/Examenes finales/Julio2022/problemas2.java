package Hector.Examenes finales.Julio2022;
import java.util.Date;


public class problemas2 {

    public class Fecha implements Comparable<Fecha> {
        private int dia, mes, anio;
    
        public Fecha() {
            Date hoy = new Date();
            this.dia = hoy.getDate();
            this.mes = hoy.getMonth() + 1;
            this.anio = hoy.getYear() + 1900;
        }
    
        public Fecha(int d, int m, int a) {
            this.dia = d;
            this.mes = m;
            this.anio = a;
        }
    
        public boolean esPosterior(Fecha otra) {
            if (anio != otra.anio) return anio > otra.anio;
            if (mes != otra.mes) return mes > otra.mes;
            return dia > otra.dia;
        }
    
        @Override
        public int compareTo(Fecha otra) {
            if (this.anio != otra.anio) {
                return this.anio - otra.anio;
            } else if (this.mes != otra.mes) {
                return this.mes - otra.mes;
            } else {
                return this.dia - otra.dia;
            }
        }
    
        @Override
        public String toString() {
            return dia + "/" + mes + "/" + anio;
        }
    }
    

    public class Donante {
        private String dni;
        private double importeTotal;
        private Fecha fechaUltimaDonacion;
    
        public Donante(String dni, double importe) {
            this.dni = dni;
            this.importeTotal = importe;
            this.fechaUltimaDonacion = new Fecha(); // fecha actual
        }
    
        public String getDni() {
            return dni;
        }
    
        public double getImporte() {
            return importeTotal;
        }
    
        public Fecha getFechaUltimaDonacion() {
            return fechaUltimaDonacion;
        }
    
        public void anadirDonativo(double cantidad) {
            importeTotal += cantidad;
            fechaUltimaDonacion = new Fecha(); // fecha actual
        }
    }
    
    public class RegistroDonantes {

        private class Nodo {
            Donante donante;
            Nodo sig;
    
            Nodo(Donante elDonante, Nodo elSiguiente) {
                this.donante = elDonante;
                this.sig = elSiguiente;
            }
        }
    
        private Nodo primero;

        void añadir(String dni, double importe){
            Nodo anterior = null;
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.donante.getDni().equals(dni)){
                    actual.donante.anadirDonativo(importe);
                    if(anterior != null){
                        anterior.sig = actual.sig;
                        actual.sig = primero;
                        primero = actual;
                    }
                    return;
                }
                anterior = actual;
            }
            primero = new Nodo(new Donante(dni, importe), primero);
        }

        void eliminarDonantesSinActividad(Fecha fecha){
            Nodo anterior = null;
            for(Nodo actual = primero; actual != null;){
                if(actual.donante.getFechaUltimaDonacion().compareTo(fecha) < 0){
                    if(anterior != null){
                        anterior.sig = actual.sig;
                        actual = actual.sig;
                    } else{
                        primero = actual.sig;
                        actual = primero;
                    }
                } else{
                    anterior = actual;
                    actual = actual.sig;
                }   
            }
        }

        void mostrarDonantesPorFecha(){
            if(primero == null){
                return;
            }
            int contador = 1;
            Nodo actual = primero;
            Fecha fechaActual = actual.donante.getFechaUltimaDonacion();
            for(; actual != null; actual = actual.sig){
                if(actual.donante.getFechaUltimaDonacion().equals(fechaActual)){
                    contador++;
                } else{
                    System.out.println(fechaActual.toString() + " " + contador + " donante(s)");
                    fechaActual = actual.donante.getFechaUltimaDonacion();
                    contador = 1;
                }
            }
            System.out.println(fechaActual.toString() + " " + contador + " donante(s)");
        }
    }
    
    
}
