package valentin;

public class Ejercicio21 {

    public static boolean validacion(int[] vector, int numero){

        for(int i = 0; i < vector.length;i++){

            if(vector[i] == numero){

                return true;
            }

        }

        return false;


    }

    public static boolean contiene(int[] primero, int[] segundo){

        for(int i = 0; i < segundo.length;i++){

            if(!validacion(primero, segundo[i])){

                return false;
            }

        }

        return true;

    }

    public static void main(String[] args) {
        int[] primero = {1,2,3,4,5,6,7,8};
        int[] segundo = {1,2,7,6,9};

        System.out.println(contiene(primero, segundo));
    }
    
}
