package Hector.Prac1;

public class Ejercicio23 {
    
    public static int contarOlasFrio(double[] tiempo, int n){

        int contador = 0;
        int olas_frio = 0;

        for(int i = 0; i < tiempo.length; i++){

            if(tiempo[i] < 0){

                contador ++;

            } else {

                if(contador >= n){

                    olas_frio ++;

                }

                contador = 0;

            }

        }

        if(contador >=n){

            olas_frio ++;

        }

        return olas_frio;

    }
    
    public static int masOlasDeFrio(double[][] matriz, int n){

        int añoOlas = -1;
        int masOlas = 0;
        
        for(int i = 0; i < matriz.length; i++){

            int olas = contarOlasFrio(matriz[i], n);

            if(olas >= masOlas){

                masOlas = olas;
                añoOlas = 1900 + i;

            }

        }

        return añoOlas;

    }

}

  


