package Hector.Examenes

import java.io.FileNotFoundException;
import java.util.Scanner;
import java.io.File;

public class problema2 {
    public class Notificar {
        public void denegar(Solicitud s) {
            System.out.println("DENEGADA: " + s.getDni() + " - " + s.getFecha());
        }
    }

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

    class ColaSolicitudes{
        private class Nodo{
            Solicitud dato;
            Nodo sig;
            
            Nodo(Solicitud s, Nodo sig){
                this.dato = s;
                this.sig = sig;
            }
        }

        private String monumento;
        private Nodo primero;
        private Nodo ultimo;

        ColaSolicitudes(String monumento){
            this.monumento = monumento;
            this.primero = null;
            this.ultimo = null;
        }

        void añadir(Solicitud s){
            Nodo nuevo = new Nodo(s, null);

            if(primero == null){
                primero = nuevo;
            } else{
                ultimo.sig = nuevo;
            }
            ultimo = nuevo;
        }

        void añadir(String fichero) throws FileNotFoundException{
            Scanner entrada = new Scanner(new File(fichero));

            while(entrada.hasNextLine()){
                String nombreMonumento = entrada.next();
                String dni = entrada.next();
                int dia = entrada.nextInt();
                int mes = entrada.nextInt();
                int año = entrada.nextInt();

                if(nombreMonumento.equals(monumento)){
                    Fecha nuevaFecha = new Fecha(dia, mes, año);
                    Solicitud nuevaSolicitud = new Solicitud(dni, nuevaFecha);
                    añadir(nuevaSolicitud);
                }
            }

            entrada.close();
        }

        String[] organizarVisita(Fecha f, int aforo){
            String[] lista = new String[aforo];
            int i =0;
            for(Nodo actual = primero; actual != null; actual = actual.sig){
                if(actual.dato.getFecha().equals(f)){
                    if(i < aforo){
                        lista[i++] = actual.dato.dni;
                    } else{
                        Notificar.denegar(actual.dato);
                    }
                }
            }

            if(i < aforo){
                String[] nuevo = new String[i];
                for(int j = 0; j < i; j++){
                    nuevo[j] = lista[j];
                }
                return nuevo;
            }

            return lista;
        }
        

    }
}
