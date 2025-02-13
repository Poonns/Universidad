package valentin;

public class Ejercicio18 {

    public static Boolean esVocal(char caracter){

        return "aeiouAEIOU".indexOf(caracter) >= 0 ? true: false;

    }

    public static boolean encaja(String patron, String cadena){

        if(patron.length() != cadena.length()){

            return false;

        }else{

            boolean resp = true;

            for(int i = 0; i< patron.length();i++){

                boolean vocal = esVocal(cadena.charAt(i));

                if(patron.charAt(i) == '.' && vocal  != true || patron.charAt(i) != '.' && patron.charAt(i) != cadena.charAt(i)){

                    return false;

                }

            }

            return resp;

        }

    }

    public static String[] add_string(String[] vector, String cadena){

        String[] nuevo_vector = new String[vector.length+1];

        for(int i = 0; i < vector.length;i++){

            nuevo_vector[i] = vector[i];
        }

        nuevo_vector[nuevo_vector.length-1] = cadena;

        return nuevo_vector;

    }


    public static String[] encajan_vectores(String[] vector, String patron){


        String[] nueva_lista = new String[0];

        for(int i = 0; i < vector.length;i++){

            if(encaja(patron, vector[i])){

                nueva_lista = add_string(nueva_lista,vector[i]);

            }

        }

        return nueva_lista;

    }

    public static void main(String[] args) {

        System.out.println("Test 1: Coincidencia exacta");
        String[] vector = {"hola", "holaa", "hola", "alo", "hol."};

        String patron = "hola";

        String[] respuesta = encajan_vectores(vector, patron);

        for(int i = 0; i < respuesta.length; i++){

            System.out.println(respuesta[i]);
        }

        System.out.println("Test 2: Con carácter comodín '.'");
        String[] vector2 = {"hola", "hila", "hula", "halo", "hero"};
        String patron2 = "h.la";

        String[] respuesta2 = encajan_vectores(vector2, patron2);

        for(int i = 0; i < respuesta2.length; i++){

            System.out.println(respuesta2[i]);
        }

        System.out.println("Test 3: Sin coincidencias");
        String[] vector3 = {"casa", "perro", "gato", "sol"};
        String patron3 = "luna";

        String[] respuesta3 = encajan_vectores(vector3, patron3);

        for(int i = 0; i < respuesta3.length; i++){

            System.out.println(respuesta3[i]);
        }

        System.out.println("Test 4: Varias vocales con '.'");
        String[] vector4 = {"pato", "peto", "pito", "poto", "pito"};
        String patron4 = "p.t.";

        String[] respuesta4 = encajan_vectores(vector4, patron4);

        for(int i = 0; i < respuesta4.length; i++){

            System.out.println(respuesta4[i]);
        }
        
    }
    
}
