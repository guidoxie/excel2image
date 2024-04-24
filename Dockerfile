FROM centos:7.6.1810

MAINTAINER guidoxie@163.com

ENV USER excel2image
ENV WORK_DIR_PATH /home/excel2image
RUN groupadd -r $USER && useradd -r -g $USER $USER
RUN mkdir -p $WORK_DIR_PATH && chown -R $USER:$USER $WORK_DIR_PATH

ENV TZ=Asia/Shanghai

COPY ["./resource/gosu", "/bin/"]
COPY ["./resource/wkhtmltox-0.12.6-1.centos7.x86_64.rpm", "/tmp/"]
COPY ["./resource/CentOS-Base.repo", "/etc/yum.repos.d/"]
COPY ["./excel2image", "/home/excel2image/"]
COPY ["./docker-entrypoint.sh", "/docker-entrypoint.sh"]
COPY ["./resource/*.TTC", "/usr/share/fonts/"]

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN set -eux; \
    yum clean all; \
    yum makecache; \
    echo "${TZ}" > /etc/timezone; \
    ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime; \
    yum install -y wget; \
    cd /tmp/ && wget --no-check-certificate https://mirrors.ustc.edu.cn/tdf/libreoffice/stable/7.6.6/rpm/x86_64/LibreOffice_7.6.6_Linux_x86-64_rpm.tar.gz; \
    yum install -y ./wkhtmltox-0.12.6-1.centos7.x86_64.rpm && tar zxvf LibreOffice_7.6.6_Linux_x86-64_rpm.tar.gz && yum install -y ./LibreOffice_7.6.6.3_Linux_x86-64_rpm/RPMS/*.rpm && rm -rf wkhtmltox-0.12.6-1.centos7.x86_64.rpm LibreOffice*;\
    yum install -y cairo cups; \
    gosu nobody true; \
    yum clean all

WORKDIR $WORK_DIR_PATH

ENTRYPOINT ["/tini", "--"]

EXPOSE 12128

CMD ["/docker-entrypoint.sh"]