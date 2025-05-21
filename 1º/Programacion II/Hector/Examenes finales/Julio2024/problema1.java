package Hector.Examenes finales.Julio2024;

public class problema1 {
    class Ruta{
        private String nombre;
        private double distancia;
        private int dificultad;

        Ruta(String nombre, double distancia, int dificultad){
            this.nombre = nombre;
            this.distancia = distancia;
            this.dificultad = dificultad;
        }

        String getNombre(){
            return this.nombre;
        }

        double getDistancia(){
            return this.distancia;
        }

        int getDificultad(){
            return this.dificultad;
        }
    }

    class CatalogoRutas{
        private Ruta[] vector;
        private static final int tamaño_inicial = 10;
        private int cantidad = 0;

        CatalogoRutas(){
            this.vector = new Ruta[tamaño_inicial];
            this.cantidad = 0;
        }

        void añadirNuevaRuta(String nombreRuta, double distancia, int dificultad){
            if(cantidad == 0){
                vector[0] = new Ruta(nombreRuta, distancia, dificultad);
                cantidad++;
                return;
            } 
            
            int pos = -1;
            for(int i = 0; i < cantidad; i++){
                if(nombreRuta.compareTo(vector[i].nombre) < 0){
                    pos = i;
                    break;
                }
            }
            if(pos == -1){
                pos = cantidad;
            }

            if(cantidad == vector.length){
                Ruta[] nuevo = new Ruta[vector.length *2];
                for(int i = 0; i < pos; i++){
                    nuevo[i] = vector[i];
                }

                nuevo[pos] = new Ruta(nombreRuta, distancia, dificultad);
                

                for(int i= pos; i < cantidad; i++){
                    nuevo[i+1] = vector[i];
                }
                vector = nuevo;
            } else{
                for(int i = cantidad; i > pos; i--){
                    vector[i] = vector[i-1];
                }
                vector[pos] = new Ruta(nombreRuta, distancia, dificultad);
            }
            cantidad++;
        }

        String[] rutasGrado(int dificultad){
            int contador = 0;
            for(int i = 0; i < cantidad; i++){
                if(vector[i].getDificultad() == dificultad){
                    contador ++;
                }
            }

            if(contador == 0){
                String[] nuevo = new String[0];
                return nuevo;
            } else{
                String[] nuevo = new String[contador];
                int j = 0;
                for(int i= 0; i < cantidad; i++){
                    if(vector[i].getDificultad() == dificultad){
                        int pos = j;
                        for(int k = 0; k < j; k++){
                            if(nuevo[k].compareTo(vector[i].nombre) > 0){
                                pos = k;
                                break;
                            }
                        }
                        
                        for(int k = j; k > pos; k--){
                            nuevo[k] = nuevo[k-1];
                        }

                        nuevo[pos] = vector[i].getNombre();
                        j++;
                    }  
                }
                return nuevo;
            }
        }

        CatalogoRutas unir(CatalogoRutas otroCatalogo){
            int i = 0;
            int j = 0;
            int k = 0;

            Ruta[] nuevo = new Ruta[vector.length + otroCatalogo.vector.length];
            while(i < cantidad && j < otroCatalogo.cantidad){
                if(vector[i].nombre.compareTo(otroCatalogo.vector[j].nombre) < 0){
                    nuevo[k++] = vector[i++];
                } else{
                    nuevo[k++] = otroCatalogo.vector[j++];
                } 
            }

            while(i < cantidad){
                nuevo[k++] = this.vector[i++];
            }

            while(j < otroCatalogo.cantidad){
                nuevo[k++] = otroCatalogo.vector[j++];
            }

            CatalogoRutas nuevoCatalogo = new CatalogoRutas();
            nuevoCatalogo.vector = nuevo;
            nuevoCatalogo.cantidad = k;

            return nuevoCatalogo;
        }
    }
}
