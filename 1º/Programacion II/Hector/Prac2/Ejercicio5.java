package Hector.Prac2;

public class Ejercicio5 {

    public class Fecha{

        private int dia;
        private int mes;
        private int año;

        public Fecha(int dia, int mes, int año){

            this.dia = dia;
            this.mes = mes;
            this.año = año;

        }

        public Fecha(Fecha otraFecha){

            this.dia = otraFecha.dia;
            this.mes = otraFecha.mes;
            this.año = otraFecha.año;

        }

        public String toString(){
            return dia + "/" + mes + "/" + año;
        }

        public boolean equals(Object otroObjeto){

            if(this == otroObjeto){
                return true;
            }

            if(otroObjeto instanceof Fecha){
                return false;
            }

            Fecha otraFecha = (Fecha) otroObjeto;
            return dia == otraFecha.dia && mes == otraFecha.mes && año == otraFecha.año;

        }

        public int compareTo(Fecha otraFecha){

            

        }
    }
    
}
