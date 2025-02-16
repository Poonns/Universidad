package Hector.Prac1;

public class Ejercicio15 {
    public static boolean estaOrdenado(String[] cadenas){

        for(int i = 1; i < cadenas.length; i++){

            if(cadenas[i].compareTo(cadenas[i-1]) > 0){
              return false;  
            }

        }

        return true;

    }
}
