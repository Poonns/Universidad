package valentin;

public class Ejercicio15 {


    public static Boolean estaOrdenado(String[] cadenas){

        Boolean resp = true;

        for(int i = 0; i < cadenas.length;i++){


            if(i+1 < cadenas.length){

                int minLen = Math.min(cadenas[i].length(), cadenas[i+1].length());
                
                for(int j = 0; j < minLen;j++){
                    if(cadenas[i].charAt(j) < cadenas[i+1].charAt(j)){

                        resp = true;
                        break;
                        
                    }else if(cadenas[i].charAt(j) > cadenas[i+1].charAt(j)){
                        
                        resp = false;
                        break;
                    }
                }
            }

            if(resp == false){
                break;
            }


        }

        return resp;
    }

    public static void main(String[] args) {

        String[] lista1 = {"amarillo", "azul", "blanco", "negro", "rojo"};//true
        String[] lista2 = {"apple", "banana", "cherry", "date"};//true
        String[] lista3 = {"rojo", "amarillo", "azul"};//false
        String[] lista4 = {"banana", "apple", "cherry"}; //false
        String[] lista5 = {"Ana", "Carlos", "Daniel", "Elena", "Roberto"};//true
        String[] lista6 = {"casa", "castillo", "caseta"}; //false
        String[] lista7 = {}; //true
        String[] lista8 = {"único", "último", "útil"}; //false
        String[] lista9 = {"rojo", "amarillo", "azul"}; //false
        String[] lista10 = {"banana", "apple", "cherry"}; //false
        String[] lista11 = {"casa", "castillo", "caseta"}; //false
        String[] lista12 = {"zebra", "ant", "dog"}; //false
        String[] lista13 = {"Ana", "elena", "Carlos"}; //false
        String[] lista14 = {"30", "10", "2", "400"};//false
        String[] lista15 = {"perro", "perrón", "perrito"};//false

        System.out.println(estaOrdenado(lista1));
        System.out.println(estaOrdenado(lista2));
        System.out.println(estaOrdenado(lista3));
        System.out.println(estaOrdenado(lista4));
        System.out.println(estaOrdenado(lista5));
        System.out.println(estaOrdenado(lista6));
        System.out.println(estaOrdenado(lista7));
        System.out.println(estaOrdenado(lista8));
        System.out.println(estaOrdenado(lista9));
        System.out.println(estaOrdenado(lista10));
        System.out.println(estaOrdenado(lista11));
        System.out.println(estaOrdenado(lista12));
        System.out.println(estaOrdenado(lista13));
        System.out.println(estaOrdenado(lista14));
        System.out.println(estaOrdenado(lista15));

        

        
    }
    
}
