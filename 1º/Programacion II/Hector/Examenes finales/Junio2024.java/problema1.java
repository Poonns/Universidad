package Hector.Examenes finales.Junio2024.java;

public class problema1 {
    public class Fecha implements Comparable<Fecha> {
        private int dia;
        private int mes;
        private int anio;
    
        public Fecha(int dia, int mes, int anio) {
            this.dia = dia;
            this.mes = mes;
            this.anio = anio;
        }
    
        // Getters
        public int getDia() {
            return dia;
        }
    
        public int getMes() {
            return mes;
        }
    
        public int getAnio() {
            return anio;
        }
    
        // compareTo para comparar fechas
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
            return String.format("%02d/%02d/%04d", dia, mes, anio);
        }
    }
    
    
    class Solicitud{
        private String dni;
        private Fecha fecha;
        
        Solicitud(String dni, Fecha f){
            this.dni = dni;
            this.fecha = f;
        }

        String getDni(){
            return this.dni;
        }

        Fecha getFecha(){
            return this.fecha;
        }
    }
    
    public class Deportista {
        private String dni;
        private String deporte;
        private Fecha fechaInicio;
    
        public Deportista(String dni, String deporte, Fecha fechaInicio) {
            this.dni = dni;
            this.deporte = deporte;
            this.fechaInicio = fechaInicio;
        }
    
        public String getDni() {
            return dni;
        }
    
        public String getDeporte() {
            return deporte;
        }
    
        public Fecha getFechaInicio() {
            return fechaInicio;
        }
    
        @Override
        public String toString() {
            return dni + " - " + deporte + " - " + fechaInicio;
        }
    }
    
    public class DeportistasAltoNivel {
        private String comunidad;
        private Deportista[] vector;
    
        public DeportistasAltoNivel(String comunidad) {
            this.comunidad = comunidad;
            this.vector = new Deportista[0];
        }
        
        String obtenerDeporte(String dni){
            int inicio = 0;
            int fin = vector.length-1;

            while(inicio <= fin){
                int media = (inicio + fin) / 2;
                int comparacion = dni.compareTo(vector[media].getDni());

                if(comparacion == 0){
                    return vector[media].getDeporte();
                } else if(comparacion < 0){
                    fin = media -1;
                } else{
                    inicio = media +1;
                }
            }
            return null;
        }

        int eliminarObsoletos(Fecha limite){
            int contador = 0;
            for(int i = 0; i < vector.length; i++){
                if(vector[i].getFechaInicio().compareTo(limite) < 0){
                    contador ++;
                }
            }
            Deportista[] nuevo = new Deportista[vector.length - contador];
            int j = 0;
            for(int i = 0; i < vector.length; i++){
                if(vector[i].getFechaInicio().compareTo(limite) >= 0){
                    nuevo[j++] = vector[i]; 
                }
            }

            vector = nuevo;

            return contador;
        }

    }
    
}
