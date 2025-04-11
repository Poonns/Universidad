package Hector.Prac2;

import java.util.Locale;

    
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
}
