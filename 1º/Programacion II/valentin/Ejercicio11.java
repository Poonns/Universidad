package valentin;

import java.util.Scanner;

public class Ejercicio11 {

    public static Boolean esVocal(char caracter){

        Boolean vocal = false;

        int[] vocales = {65,69,73,79,85};

        for(int i = 0; i < vocales.length;i++){

            if(vocales[i] == (int) caracter || vocales[i]+32 == (int) caracter){

                vocal = true;
                break;
            }

        }

        return vocal;

    }




    public static Boolean encaja(String patron, String cadena){

        if(patron.length() != cadena.length()){

            return false;

        }else{

            Boolean resp = true;

            for(int i = 0; i< patron.length();i++){

                Boolean vocal = esVocal(cadena.charAt(i));

                if(patron.charAt(i) == '.' && vocal  != true || patron.charAt(i) != '.' && patron.charAt(i) != cadena.charAt(i)){

                    return false;

                }

            }

            return resp;

        }

    }


    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);

        System.out.println("Escribe el patron: ");
        
        String patron = sc.nextLine();
        
        System.out.println("Escribe el patron: ");

        String cadena = sc.nextLine();

        sc.close();

        System.out.println(encaja(patron,cadena));
    }
    
}
