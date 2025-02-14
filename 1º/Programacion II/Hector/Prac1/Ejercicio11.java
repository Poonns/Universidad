package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio11 {

    public static boolean esVocal(char c) {
        return "AEIOUaeiou".indexOf(c) >= 0;

        }
    
    public static boolean encaja(String patron, String cadena){

        if(cadena.length() != patron.length()){

    }

    public static boolean encaja(String patron, String cadena) {
        if (cadena.length() != patron.length()) {

            return false;
        }
         
        for(int i = 0; i < cadena.length(); i++){


            char patron_caracteres = patron.charAt(i);
            char cadena_caracteres = cadena.charAt(i);

            if(patron_caracteres == '.'){

                if(!esVocal(cadena_caracteres)){
                        return false;
                }

            } else if(patron_caracteres != cadena_caracteres){
                return false;
            }

        }

        return true;
    }
          


    

        for (int i = 0; i < cadena.length(); i++) {
            char patron_caracter = patron.charAt(i);
            char cadena_caracter = cadena.charAt(i);

            if (patron_caracter == '.') {
                if (!esVocal(cadena_caracter)) {
                    return false;
                }
            } else if (patron_caracter != cadena_caracter) {
                return false;
            }
        }
        return true;
    }


    public static void main(String[] args) {
        Scanner entrada = new Scanner(System.in);
        System.out.println("Introduce un patron: ");
        String cadena1 = entrada.nextLine();
        System.out.println("Introduce otra cadena: ");
        String cadena2 = entrada.nextLine();
        entrada.close();

        System.out.println(encaja(cadena1, cadena2));
    }
}