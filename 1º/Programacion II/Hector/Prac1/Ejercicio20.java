package Hector.Prac1;

public class Ejercicio20 {

    public static int[] eliminarValor(int[] enteros, int n){

        int contador = 0;

        for(int elemento : enteros){

            if(elemento == n){

                contador ++;

            }

        }

        int[] resultado = new int[enteros.length - contador];

         int indice = 0;

        for(int elemento : enteros){

            if(elemento != n){

                resultado[indice] = elemento;
                indice ++;

            }

        }

        return resultado;

    }
    
}
