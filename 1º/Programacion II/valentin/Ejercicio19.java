package valentin;

import java.util.Scanner;

public class Ejercicio19 {

    public static int[] eliminarPosicion(int[] arreglo,int index){

        if(index <= 0 || index > arreglo.length){

            return arreglo;
        }

        int[] nuevo_arreglo = new int[arreglo.length-1];

        int cont = 0;

        for(int i = 0; i < arreglo.length;i++){

            if(i != index-1){

                nuevo_arreglo[cont] = arreglo[i];

                cont+=1;

            }

        }

        return nuevo_arreglo;

    }

    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);

        System.out.println("Dime una posicion del vector para eliminar el numero: ");
        int index = sc.nextInt();


        sc.close();

        int[] arreglo = {1,2,3,4,5,6,7,8,9};

        int[] respuesta = eliminarPosicion(arreglo,index);

        for(int i = 0; i < respuesta.length;i++){

            System.out.print(respuesta[i]);
        }
    }
    
}
