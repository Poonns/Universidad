package Hector.Prac1;

public class Ejercicio19 {
    
    public static int[] eliminarPosicion(int[] enteros, int posicion){

        if(posicion < 0 || posicion >= enteros.length){

            return enteros;

        }

        int[] resultado = new int[enteros.length-1];

        int indice = 0;

        for(int i = 0; i < enteros.length; i++){

            if(i != posicion){

                resultado[indice] = enteros[i];
                indice ++;             
            }

        }

    return resultado;

    }
}
