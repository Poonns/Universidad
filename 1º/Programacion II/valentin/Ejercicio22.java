package valentin;

public class Ejercicio22 {


    public static int verificacion(int[] vector, int n){

        for(int i = 0; i < vector.length; i++){

            if(n == vector[i]){

                return i;

            }
            
        }
        return -1;

    }

    public static int posicionInsercion(int[] vector, int numero){

        int verificacion = verificacion(vector, numero);

        if(verificacion != -1){

            return verificacion;
        }

        for(int i = 0; i < vector.length; i++){

            if(i == 0 && numero < vector[i+1]){

                return i;

            }else if(i == vector.length-1 && numero > vector[i]){

                return i+1;
                
            }else if(i == vector.length-1 && numero < vector[i]){

                return i;
                
            }else if(i != 0 && numero > vector[i-1] && numero < vector[i+1]){

                return i+1;

            }

        }

        return -1;
    }

    public static void main(String[] args) {

        int[] vector = {1,2,3,4,5,6,7,8,9};

        int numero = 11;

        System.out.println(posicionInsercion(vector, numero));
        
    }
    
}
