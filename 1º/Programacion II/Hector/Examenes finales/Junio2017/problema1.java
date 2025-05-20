package Hector.Examenes finales.Junio2017;

public class problema1 {
    class Compra{
        private String dni;
        private int cantidad;

        Compra(String dni, int cantidad){
            this.dni = dni;
            this.cantidad = cantidad;
        }

        String getDni(){
            return this.dni;
        }
        
        int getCantidad(){
            return this.cantidad;
        }

        void setCantidad(int cantidad){
            this.cantidad = cantidad;
        }
    }

    class ComprasEspectaculo{
        private final int tamaño_inicial = 10;
        private String descripcion;
        private int maximo;
        private Compra[] compras;
        private int ocupados;

        ComprasEspectaculo(String descripcion, int maximo){
            this.descripcion = descripcion;
            this.maximo = maximo;
            this.compras = new Compra[tamaño_inicial];
            this.ocupados = 0;
        }

        int consultarEntradas(String dni){
            int i = -1;
            int inicio = 0;
            int fin = compras.length -1;

            while(inicio <= fin){
                 int media = (inicio + fin) / 2;
                 String actual = compras[media].getDni();
                 int comparacion = dni.compareTo(actual);

                 if(comparacion == 0){
                    i = media;
                 } else if(comparacion < 0){
                    fin = media -1;
                 } else{
                    inicio = media +1;
                 }
            }

            if(i == -1){
                return 0;
            } else {
                return compras[i].cantidad;
            }
        }

        private void redimensionar(int nuevoTamaño) throws IllegalArgumentException{
            if(nuevoTamaño < ocupados){
                throw new IllegalArgumentException("Tamaño no valido");
            } else{
                Compra[] nuevoVector = new Compra[nuevoTamaño];
                for(int i = 0; i < ocupados; i++){
                    nuevoVector[i] = compras[i];
                }
                compras = nuevoVector;
            }
        }

        boolean añadirCompra(String dni, int cantidad){
            for(int i = 0; i < ocupados; i++){
                if(compras[i].getDni().equals(dni)){
                    int total = compras[i].getCantidad() + cantidad;
                    if(total <= maximo){
                        return true;
                    } else {
                        return false;
                    }
                }
            }
            
            if(ocupados == compras.length){
                redimensionar(compras.length * 2);
            }

            int pos = 0;
            for(int i = 0; i < ocupados; i++){
                if(compras[i].getDni().compareTo(dni) > 0){
                    pos = i;
                }
            }

            for(int i = ocupados; i < pos; i--){
                compras[i] = compras[i-1];
            }

            compras[pos] = new Compra(dni, cantidad);
            ocupados ++;

            return true;
        }

        int contarComunes(ComprasEspectaculo otro){
            int contador = 0;
            int i = 0;
            int j = 0;

            while(i < ocupados && j < otro.ocupados){
                String dni1 = compras[i].getDni();
                String dni2 = otro.compras[j].getDni();

                int comparacion = dni1.compareTo(dni2);
                if(comparacion == 0){
                    contador ++;
                    i++;
                    j++;
                } else if(comparacion < 0){
                    i++;
                } else {
                    j++;
                }
            }

            return contador;
        }
    }
}
