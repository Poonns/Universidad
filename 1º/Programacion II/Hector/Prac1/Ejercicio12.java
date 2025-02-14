package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio12 {

    public static int contarPalabras(String cadena){

        int contador_palabras = 0;

        for(int i = 0; i < cadena.length(); i++){

            char caracteres_cadena = cadena.charAt(i);

            if(caracteres_cadena != ' ' && (i == 0 || cadena.charAt(i - 1) == ' ')){

                contador_palabras ++;

            }

        }

        return contador_palabras;

    }

    public static void main(String[] args) {
        
        Scanner entrada = new Scanner(System.in);
        System.out.println("Introduce una cadena: ");
        String cadena = entrada.nextLine();
        entrada.close();

        System.out.println(contarPalabras(cadena));



    }
    
}
