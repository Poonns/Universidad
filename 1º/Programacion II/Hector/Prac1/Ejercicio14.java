package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio14 {
    
    public static int ultimaPos(int[] enteros , int n){

        int pos = -1;

        for(int i = 0; i < enteros.length; i++){

            if(enteros[i] == n){

                pos = i;

            } 

        }

        return pos;

    }


    public static void main(String[] args) {
        
        Scanner entrada = new Scanner(System.in);
        int[] enteros = {1,2,3,4,5};
        System.out.println("Introduce un numero que quieras buscar en el vector: ");
        int vector = entrada.nextInt();
        entrada.close();

        System.out.println(ultimaPos(enteros, vector));


    }








}
