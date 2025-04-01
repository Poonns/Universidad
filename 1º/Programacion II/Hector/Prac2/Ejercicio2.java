package Hector.Prac2;

import java.util.Locale;
import java.util.Scanner;
import java.io.File;
import java.io.FileNotFoundException;

public class Ejercicio2 {

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

    public class Restaurante{

        private String restaurante;
        private Punto posicion;
        private int valoracion;

        public Restaurante(String nombre, Punto posicion, int valoracion){

            this.restaurante = nombre;
            this.posicion = posicion;
            this.valoracion = valoracion;
        }

        public String getNombre(){

            return this.restaurante;
        }

        public Punto getPosicion(){

            return this.posicion;
        }

        public int getValoracion(){

            return this.valoracion;
        }

        public double distancia(Punto p){

            return posicion.distancia(p);
        }

        public Restaurante[] leeRestaurantes(String fichero) throws FileNotFoundException {

            Scanner entrada = new Scanner(new File(fichero));

            while(entrada.hasNextLine()){

                String[] partes = entrada.nextLine().split(" ");

                if(partes.length == 4){

                    int cantidad_restaurantes = partes[0];
                        
                }
            }

        }

    }
    
}
