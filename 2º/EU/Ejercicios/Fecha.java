public class Fecha implements Comparable<Fecha> {
    private int dia, mes, ano;

    public Fecha(int dia, int mes, int ano) {
        this.dia = dia;
        this.mes = mes;
        this.ano = ano;
    }

    public int compareTo(Fecha otra) {
        if (ano != otra.ano) return ano - otra.ano;
        if (mes != otra.mes) return mes - otra.mes;
        return dia - otra.dia;
    }

    public String toString() {
        return dia + "/" + mes + "/" + ano;
    }
}


public class Medicamento {
    private String codigo;
    private int cantidad;
    private Fecha fechaUltimaOperacion;

    public Medicamento(String codigo, int cantidad) {
        this.codigo = codigo;
        this.cantidad = cantidad;
        this.fechaUltimaOperacion = new Fecha();
    }

    public String getCodigo() {
        return codigo;
    }

    public int getStock() {
        return cantidad;
    }

    public Fecha getFecha() {
        return fechaUltimaOperacion;
    }

    public void modificarCantidad(int nuevaCantidad) {
        this.cantidad = nuevaCantidad;
        this.fechaUltimaOperacion = new Fecha();
    }

    @Override
    public String toString() {
        return "Código: " + codigo + ", Cantidad: " + cantidad + ", Fecha última operación: " + fechaUltimaOperacion;
    }

    class AlmacenMedicamentos{
        private static final int tam_inicial = 10;
        private Medicamento[] vector;
        private int ocupados;

        public AlmacenMedicamentos() {
            vector = new Medicamento[tam_inicial];
            ocupados = 0;
        }
        
        int getStock(String codigo){
            int inicio = 0;
            int fin = 0;

            while(inicio <= fin){
                int media = (inicio + fin) / 2;
                int comparacion = vector[media].getCodigo().compareTo(codigo);

                if(comparacion == 0){
                    return vector[media].getStock();
                } else if(comparacion < 0){
                    inicio = media +1;
                } else{
                    fin = media -1;
                }
            }
            return -1;
        }

        void eliminar(Fecha f){
            int contador = 0;
            for(int i = 0; i < ocupados; i++){
                if(vector[i].getFecha().compareTo(f) < 0){
                    contador ++;
                }
            }

            Medicamento[] nuevo = new Medicamento[tam_inicial - contador];
            int j = 0;
            for(int i = 0; i < ocupados; i++){
                if(vector[i].getFecha().compareTo(f) >= 0){
                    nuevo[j] = vector[i];
                    j++;
                }
            }
            vector = nuevo;
            ocupados = j;
        }

        AlmacenMedicamentos unir(AlmacenMedicamentos otro){
            Medicamento[] nuevo = new Medicamento[ocupados + otro.ocupados];
            int i = 0;
            int j = 0;
            int k = 0;

            while(i < ocupados && j < otro.ocupados){
                if(vector[i].getFecha().compareTo(otro.vector[j].getFecha()) < 0){
                    nuevo[k++] = new Medicamento(vector[i].codigo, vector[i++].cantidad);
                } else if(vector[i].getFecha().compareTo(otro.vector[j].getFecha()) > 0){
                    nuevo[k++] = new Medicamento(vector[j].codigo, vector[j++].cantidad); 
                } else{
                    int suma = vector[i].cantidad += otro.vector[j].cantidad;
                    nuevo[k++] = new Medicamento(vector[j].codigo, suma);
                }
            }

            while(i < ocupados){
                nuevo[k++] = new Medicamento(vector[i].codigo, vector[i++].cantidad);
            }

            while(j < otro.ocupados){
                nuevo[k++] = new Medicamento(vector[j].codigo, vector[j++].cantidad);
            }

            AlmacenMedicamentos nuevoAlmacen = new AlmacenMedicamentos();
            this.vector = nuevo;
            ocupados = k;
            return nuevoAlmacen;
        }
    }
}