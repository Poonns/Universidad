package Hector.Prac1;

public class Ejercicio18 {
    
    public static boolean esVocal(char c) {
        return "AEIOUaeiou".indexOf(c) >= 0;

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
    
    
    
    public static String[] encajan(String patron, String[] cadenas){

        int contador = 0;
        
        for(String cadena : cadenas){

            if(encaja(patron, cadena)){

                contador ++;

            }

        }

        String[] resultado = new String[contador];

        int pos = 0;

        for(String cadena : cadenas){

            if(encaja(patron, cadena)){

                resultado[pos] = cadena;
                pos ++;

            }

    }

    return resultado;

    }

}


