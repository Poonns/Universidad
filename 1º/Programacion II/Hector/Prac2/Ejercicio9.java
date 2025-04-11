package Hector.Prac2;

import java.util.Locale;
import java.util.Vector;

public class Ejercicio9 {
    public class Punto{

        private double x;
        private double y;

        public Punto(double x, double y){

                this.x = x;
                this.y = y;

        }

        public Punto(){

            this.x = 0;
            this.y = 0;

        }

        public double getX(){

            return x;

        }

        public double getY(){

            return y;

        }

        public Punto desplazar(double desplazaminetoX, double desplazamientoY){

            Punto nuevo_punto = new Punto(this.x + desplazaminetoX, this.y + desplazamientoY);
            return nuevo_punto;

        }

        public double distancia(Punto otroPunto){

            double distancia = Math.sqrt(Math.pow(otroPunto.getX() - this.x, 2) + Math.pow(otroPunto.getY() - this.y, 2));
            return distancia;

        }

        public String toString(){

            return String.format(Locale.US, "( %.2f, %.2f)", x, y);

        }

        public boolean equals(Object otroObjeto){
            
            if(this == otroObjeto){
                return true;
            }

            if(otroObjeto == null || getClass() != otroObjeto.getClass()){
                return false;
            }

            Punto otroPunto = (Punto) otroObjeto; //'cast' hacemos un objeto mas general en uno mas específico
            return this.x == otroPunto.x && this.y == otroPunto.y;
        }  
    }
    
    public class LineaPoligonal{
        private Punto[] puntos;
        private int cantidad;

        public LineaPoligonal(){
            puntos = new Punto[10];
            cantidad = 0;
        }

        public void añadir(Punto punto){

            if(cantidad == puntos.length){
                Punto[] nuevoVector = new Punto[puntos.length *2];
                for(int i = 0; i < puntos.length; i++){
                    nuevoVector[i] = puntos[i];
                }
                puntos = nuevoVector;
                
            }
            puntos[cantidad] = punto;
            cantidad ++;
        }

        public void quitar(int posicion){
            for(int i = posicion; i < cantidad -1; i++){
                puntos[i] = puntos[i+1];
            }
            puntos[cantidad-1] = null;
            cantidad --;
        }

        public void quitar(Punto punto){
            for(int i = 0; i < cantidad; i++){
                if(puntos[i].equals(punto)){
                    quitar(i);
                }
            }
        }

        public void trasladar(double desplazamientoX, double desplazamientoY){
            for(int i = 0; i < cantidad; i++){
                puntos[i] = puntos[i].desplazar(desplazamientoX, desplazamientoY);
            }
        }

        public double longitud(){
            if(cantidad == 0 || cantidad == 1){
                return 0;
            }

            int distancias = 0;
            for(int i = 0; i < cantidad; i++){
                if(cantidad > 1){
                    distancias += puntos[i].distancia(puntos[i+1]);
                }
            }
            return distancias;
        }
    }
}
    

