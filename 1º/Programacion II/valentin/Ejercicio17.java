package valentin;

public class Ejercicio17 {
    public static Boolean hayRepetidos(int[] vector){
        for(int i = 0;i < vector.length; i++){

            for(int j = 0; j < vector.length;j++){

                if(i == j){
                    
                    continue;

                }else if(vector[i] == vector[j]){

                    return false;

                }

            }

        }

        return true;
    }

    public static void main(String[] args) {
        int[] numeros1 = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10};//true
        int[] numeros2 = {7, 7, 7, 7, 7, 7, 7, 7};//false
        int[] numeros3 = {4, 1, 6, 2, 8, 4, 9, 6, 3, 5, 2};//false
        int[] numeros4 = {11, 22, 33, 44, 55, 11, 22, 33, 44, 55};//false
        int[] numeros5 = {15, 27, 39, 42, 15, 56, 39, 78, 99, 27, 100, 42, 56, 78};//false
        int[] numeros6 = {42};//true
        int[] numeros7 = {3, 3, 5, 5, 7, 7, 9, 9, 11, 11};//false
        int[] numeros8 = {-4, -2, 0, 3, 5, -2, 8, -4, 10, 3};//false
        int[] numeros9 = {2, 4, 6, 8, 10, 12, 4, 8, 14, 16, 2};//false
        int[] numeros10 = {1, 3, 5, 7, 9, 11, 13, 3, 7, 15, 17, 5};//false
        int[] numeros11 = {100, 200, 300, 400, 500, 600, 700, 800, 900, 100, 300, 500, 700, 900};//false
        int[] numeros12 = {10, 9, 8, 7, 6, 5, 4, 3, 2, 1};//true
        int[] numeros13 = {17, 34, 51, 68, 85, 99, 123, 146, 157, 176};//true
        int[] numeros14 = {};//true
        int[] numeros15 = {5, 5, 5, 5, 5, 10, 10, 10, 15, 15, 20, 20, 20, 20};//false


        System.out.println(hayRepetidos(numeros1));
        System.out.println(hayRepetidos(numeros2));
        System.out.println(hayRepetidos(numeros3));
        System.out.println(hayRepetidos(numeros4));
        System.out.println(hayRepetidos(numeros5));
        System.out.println(hayRepetidos(numeros6));
        System.out.println(hayRepetidos(numeros7));
        System.out.println(hayRepetidos(numeros8));
        System.out.println(hayRepetidos(numeros9));
        System.out.println(hayRepetidos(numeros10));
        System.out.println(hayRepetidos(numeros11));
        System.out.println(hayRepetidos(numeros12));
        System.out.println(hayRepetidos(numeros13));
        System.out.println(hayRepetidos(numeros14));
        System.out.println(hayRepetidos(numeros15));
    }
}
