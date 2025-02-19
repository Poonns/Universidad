package Hector.Prac1;

public class Ejercicio21 {

    public static boolean contiene(int[] enteros, int n){

        for(int elemento : enteros){

            if(elemento == n){

                return true;

            }

        }

        return false;

    }

    public static boolean contiene(int[] v1, int[] v2){

        for(int elemento : v2){

            if(!contiene(v1, elemento)){

                return false;

            }

        }

        return true;

    }
    
}
