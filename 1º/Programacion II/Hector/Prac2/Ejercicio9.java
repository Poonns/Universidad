package Hector.Prac2;

public class Ejercicio9 {

    public class Punto{

        private int x;
        private int y;

        public Punto(int x, int y){
            this.x = x;
            this.y = y;
        }

        public int getX(){
            return this.x;
        }

        public int getY(){
            return this.y;
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

            if(cantidad >= puntos.length){
                Punto[] nuevoVector = new Punto[puntos.length *2 + 1];
                for(int i = 0; i < puntos.length; i++){
                    nuevoVector[i] = puntos[i];
                }
            }
        }
    }
    
}
