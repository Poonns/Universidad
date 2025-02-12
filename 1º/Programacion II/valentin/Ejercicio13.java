package valentin;

import java.util.Scanner;

public class Ejercicio13 {

    public static int contarPalabras(String cadena){

        int caracteres = 0;

        int respuesta = 0;

        for(int i = 0; i < cadena.length(); i++){

            int caracter_ascii = (int)cadena.charAt(i);

            if(caracter_ascii == 32 && caracteres != 0){
                 respuesta+=1;
                 caracteres = 0;
            }else if(caracter_ascii != 32){
                caracteres+=1;
            }

        }

        if(caracteres != 0){
            respuesta+=1;
        }

        return respuesta;

    }


    public static String obtenerPalabra(String cadena, int index){

        int palabras_contar = contarPalabras(cadena);

        if(palabras_contar == 0 || index <= 0 || index > palabras_contar ){

            return null;

        }else{


            String[] palabras = new String[0];

            String palabra = "";

            for(int i = 0; i < cadena.length(); i++){

                int caracter_ascii = (int)cadena.charAt(i);
    
                if(caracter_ascii == 32){
                    if(palabra != "" && palabras.length >= 1){
                        String[] palabrasLista = new String[palabras.length+1];

                        for(int j = 0; j < palabras.length; j++){

                            palabrasLista[j] = palabras[j];

                        }

                        palabrasLista[palabrasLista.length-1] = palabra;

                        palabras = palabrasLista;

                        palabra = "";

                    }else if(palabra != "" && palabras.length == 0){
                        String[] palabrasLista = new String[palabras.length+1];

                        palabrasLista[0] = palabra;

                        palabras = palabrasLista;

                        palabra = "";
                    }

                }else if(caracter_ascii != 32){
                    palabra+=cadena.charAt(i);
                }
    
            }

            if(palabra != ""){
                String[] palabrasLista = new String[palabras.length+1];

                for(int j = 0; j < palabras.length; j++){

                    palabrasLista[j] = palabras[j];

                }

                palabrasLista[palabrasLista.length-1] = palabra;

                palabras = palabrasLista;

            }

            return palabras[index-1];
        }
    }


    public static void main(String[] args) {
        
        Scanner sc = new Scanner(System.in);

        System.out.println("Escribe la cadena: ");

        String cadena = sc.nextLine();

        System.out.println("Escribe que numero de palabra quieres: ");

        int index = sc.nextInt();

        sc.close();


        System.out.println(obtenerPalabra( cadena, index));
    }
    
}
