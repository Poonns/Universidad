package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio11 {

    public static boolean encaja(String patron, String cadena){

        if(cadena.equals(patron)){
            return true;
        } else {
            return false;
        }

    }
    
    public static void main(String[] args) {
        
        Scanner entrada = new Scanner(System.in);
        System.out.println("Introduce una cadena: ");
        String cadena1 = entrada.nextLine();
        System.out.println("Introduce otra cadena: ");
        String cadena2 = entrada.nextLine();
        entrada.close();

        System.out.println(encaja(cadena1, cadena2));

    }

}
