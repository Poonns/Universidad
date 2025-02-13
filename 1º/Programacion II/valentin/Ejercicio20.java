package valentin;

import java.util.Scanner;

public class Ejercicio20 {
    public static int[] eliminarValor(int[] vector, int numero){

        int[] arreglo_respuesta = new int[0];

        for(int i = 0; i < vector.length; i++){

            if(vector[i] != numero){

                int[] nuevo_arreglo_respuesta = new int[arreglo_respuesta.length+1];


                    for(int j = 0; j < arreglo_respuesta.length; j++){

                        nuevo_arreglo_respuesta[j] = arreglo_respuesta[j];
                    }

                    nuevo_arreglo_respuesta[nuevo_arreglo_respuesta.length-1] = vector[i];

                    arreglo_respuesta = nuevo_arreglo_respuesta;


            }

        }

            return arreglo_respuesta;
    }


    public static void main(String[] args) {

        Scanner sc = new Scanner(System.in);

        int numero = sc.nextInt();

        sc.close();
    
        int[] vector = {1,2,3,4,5,1,2,3,6};

        int[] respuesta = eliminarValor(vector, numero);

        for(int i = 0; i < respuesta.length;i++){

            System.out.print(respuesta[i]+" ");
        }

    }
}
