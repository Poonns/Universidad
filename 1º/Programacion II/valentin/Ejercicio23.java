package valentin;

public class Ejercicio23 {

    public static int olasDeFrio(double[] temperaturas, int numero){

        int cont = 0;

        int resp = 0;

        for(int i = 0; i < temperaturas.length;i++){

            if(temperaturas[i] < 0){

                cont+=1;

            }else if(temperaturas[i] > 0){
               cont = 0;
            }

            if(cont == numero){

                resp+=1;
                cont = 0;
            }

        }

        return resp;
    }


    public static int masOlasDeFrio(double[][] temperaturas, int n){

        int cantidad_temperatura = 0;

        int index = 0;

        for(int i = 0; i< temperaturas.length;i++){
            int cantidad_por_año = olasDeFrio(temperaturas[i], n);

            if(cantidad_por_año >= cantidad_temperatura){

                cantidad_temperatura = cantidad_por_año;

                index = i;

            }
        }

        if(cantidad_temperatura == 0){

            return -1;
        }


        return index + 1900;
    }

    public static void main(String[] args) {


        double[][] temperaturas =  {
            { -2.5, 0.0, -1.2, 3.5, 4.8, 5.9, -6.0 },
            { 2.1, -1.4, 3.3, 5.5, -0.2, 4.4, 6.0 },
            { -2.0, -0.5, 1.0, 2.5, 3.7, -5.8, 6.0 },
            { 4.0, 3.2, 2.1, -1.0, 0.5, 1.8, 5.9 },
            { 1.1, -2.3, 3.4, 4.5, 5.6, -2.3, 6.0 },
            { -0.9, 1.5, 2.7, 3.9, 5.2, -1.8, 6.0 },
            { 0.2, -1.8, -2.9, -2.1, -4.7, 5.5, -6.0 },
            { 3.3, 2.2, 1.1, 0.0, -1.1, 4.4,-6.0 },
            { -2.5, -1.0, 0.5, 2.0, -3.5, 5.0, -6.0 },
            { 5.6, -4.4, -3.2, -2.0, 1.0, -2.0, 6.0 }
        };

        System.out.println(masOlasDeFrio(temperaturas, 3));
    }
    
}
