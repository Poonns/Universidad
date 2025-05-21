package Hector.Examenes finales.Junio2022;
import java.util.Date;

public class problema1 {

    public class Fecha implements Comparable<Fecha> {
        private int dia, mes, anio;

        public Fecha() {
            Date hoy = new Date();
            this.dia = hoy.getDate();
            this.mes = hoy.getMonth() + 1;
            this.anio = hoy.getYear() + 1900;
        }

        public Fecha(int d, int m, int a) {
            this.dia = d;
            this.mes = m;
            this.anio = a;
        }

        @Override
        public int compareTo(Fecha otra) {
            if (anio != otra.anio) return anio - otra.anio;
            if (mes != otra.mes) return mes - otra.mes;
            return dia - otra.dia;
        }

        @Override
        public boolean equals(Object obj) {
            if (obj instanceof Fecha) {
                Fecha otra = (Fecha) obj;
                return dia == otra.dia && mes == otra.mes && anio == otra.anio;
            }
            return false;
        }

        @Override
        public String toString() {
            return dia + "/" + mes + "/" + anio;
        }
    }

    public class Medicamento {
        private String codigo;
        private int cantidad;
        private Fecha fecha;
    
        public Medicamento(String codigo, int cantidad) {
            this.codigo = codigo;
            this.cantidad = cantidad;
            this.fecha = new Fecha(); // fecha actual
        }
    
        public String getCodigo() {
            return codigo;
        }
    
        public int getStock() {
            return cantidad;
        }
    
        public Fecha getFecha() {
            return fecha;
        }
    
        public void añadirStock(int cantidad) {
            this.cantidad += cantidad;
            this.fecha = new Fecha(); // actualiza a fecha actual
        }
    
        public void retirarStock(int cantidad) {
            this.cantidad -= cantidad;
            this.fecha = new Fecha(); // actualiza a fecha actual
        }
    
        @Override
        public String toString() {
            return codigo + " | Stock: " + cantidad + " | Fecha: " + fecha;
        }
    }

    public class AlmacenMedicamentos {
        private static final int TAM_INICIAL = 10;
        private Medicamento[] vector;
        private int ocupados;
    
        public AlmacenMedicamentos() {
            vector = new Medicamento[TAM_INICIAL];
            ocupados = 0;
        }
    
        int getStock(String codigo){
            int inicio = 0;
            int fin = vector.length-1;

            while(inicio <= fin){
                int media = (inicio + fin) / 2;
                String actual = vector[media].getCodigo();
                int comparacion = codigo.compareTo(actual);

                if(comparacion == 0){
                    return vector[media].getStock();
                } else if(comparacion < 0){
                    fin = media - 1;
                } else{
                    inicio = media +1;
                }
            }
            return -1;
        }

        void eliminar(Fecha f){
            int contador = 0;
            for(int i = 0; i < ocupados; i++){
                if(vector[i].getFecha().compareTo(f) < 0){
                    contador++;
                }
            }

            Medicamento[] nuevo = new Medicamento[ocupados - contador];
            int j = 0;
            for(int i = 0; i < ocupados; i++){
                if(vector[i].getFecha().compareTo(f) >= 0){
                    nuevo[j++] = vector[i];
                }
            }

            vector = nuevo;
            ocupados = j;
        }

        AlmacenMedicamentos unir(AlmacenMedicamentos otro){
            int i = 0;
            int j = 0;
            int k = 0;

            Medicamento[] nuevo = new Medicamento[ocupados + otro.ocupados];
            while(i < ocupados && j < otro.ocupados){
                if(vector[i].getCodigo().compareTo(otro.vector[j].getCodigo()) < 0){
                    nuevo[k++] = new Medicamento(vector[i].getCodigo(), vector[i++].getStock());
                } else if(vector[i].getCodigo().compareTo(otro.vector[j].getCodigo()) > 0){
                    nuevo[k++] = new Medicamento(otro.vector[j].getCodigo(), otro.vector[j++].getStock());
                } else{
                    nuevo[k++] = new Medicamento(vector[i].getCodigo(), otro.vector[j++].getStock() + vector[i++].getStock());
                }
            }

            while(i < ocupados){
                nuevo[k++] = new Medicamento(vector[i].getCodigo(), vector[i++].getStock());
            }

            while(j < otro.ocupados){
                nuevo[k++] = new Medicamento(otro.vector[j].getCodigo(), otro.vector[j++].getStock());
            }
             AlmacenMedicamentos nuevoAlmacen = new AlmacenMedicamentos();
             nuevoAlmacen.vector = nuevo;
             nuevoAlmacen.ocupados = k;

             return nuevoAlmacen;

        }
    }
}