package Hector.Prac1;

import java.util.Scanner;

public class practica {   

    public static int ultimaPosicion(int[] enteros, int n){

        for(int i = enteros.length - 1; i>= 0; i--){
            if(enteros[i] == n){
                return i;
            }
        }

        return -1;

    }

}

