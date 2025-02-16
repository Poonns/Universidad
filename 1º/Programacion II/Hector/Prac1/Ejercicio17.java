package Hector.Prac1;

public class Ejercicio17 {
    
    public static boolean hayRepetidos(int[] enteros){

        for(int i = 0; i < enteros.length; i++){

            for(int j = i + 1; j < enteros.length; j++){

                if(enteros[i] == enteros[j]){

                    return true;
                    

                }

            }

        }

        return false;

    }

}
