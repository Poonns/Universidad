package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio13 {

    public static String obtenerPalabra(String cadena, int n){

        int contador_palabras = 0;
        
        for(int i = 0; i < cadena.length(); i++){

         /*    char cadena_caracteres = cadena.charAt(i); */

            if(cadena.charAt(i) != ' ' && (i == 0 || cadena.charAt(i - 1) == ' ')){

                contador_palabras ++;

                if(contador_palabras == n){

                    int inicio = i;
                    int fin = i;

                    while(fin < cadena.length() && cadena.charAt(fin) != ' '){
                        fin++;
                    }

                    return cadena.substring(inicio, fin);

                }

            }

        }

        return null;

    }


    public static void main(String[] args) {
        
        Scanner entrada = new Scanner(System.in);
        System.out.println("Introduce una cadena: ");
        String cadena = entrada.nextLine();
        System.out.println("Introduce el numero de palabra que quieres que muestre: ");
        int n = entrada.nextInt();
        entrada.close();


        System.out.println(obtenerPalabra(cadena, n));


    }








    
}
