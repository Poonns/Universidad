package valentin;

import java.util.Scanner;

public class Ejercicio12 {

    public static int contarPalabras(String cadena){

        int caracteres = 0;

        int respuesta = 0;

        for(int i = 0; i < cadena.length(); i++){

            int caracter_ascii = (int)cadena.charAt(i);

            if(caracter_ascii == 32 && caracteres != 0){
                 respuesta+=1;
                 caracteres = 0;
            }else{
                caracteres+=1;
            }

        }

        if(caracteres != 0){
            respuesta+=1;
        }

        return respuesta;

    }

    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);

        
        System.out.println("Escribe la cadena: ");

        String cadena = sc.nextLine();

        sc.close();

        System.out.println(contarPalabras(cadena));
    }
    
}
