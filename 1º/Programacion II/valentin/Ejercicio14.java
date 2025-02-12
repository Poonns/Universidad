package valentin;

import java.util.Scanner;

public class Ejercicio14 {

    public static int ultimaPosicion(int[] numeros_vector, int numero){

        int index = -1;

        for(int i = 0; i < numeros_vector.length;i++){

            if(numeros_vector[i] == numero){

                index = i;

            }

        }

        return index;

    }



    public static void main(String[] args) {

        Scanner sc = new Scanner(System.in);

        int num =  sc.nextInt();

        sc.close();

        int[] numeros_vector = {3,5,2,5,8,2,5};

        System.out.println(ultimaPosicion(numeros_vector,num));


    }
    
}
