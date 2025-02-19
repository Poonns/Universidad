package Hector.Prac1;

public class Ejercicio22 {
    
    public static int posicionInsercion(int[] enteros, int n){


        for (int i = 0; i < enteros.length; i++){

            if(enteros[i] >= n){

                return i;

            }

        }

        return enteros.length;

    }

}
